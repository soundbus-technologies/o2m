// authors: wangoo
// created: 2018-05-30
// oauth2 user store

package o2m

import (
	"gopkg.in/mgo.v2"
	"github.com/go2s/o2x"
	"reflect"
	"gopkg.in/mgo.v2/bson"
	"github.com/golang/glog"
	"github.com/patrickmn/go-cache"
	"time"
	"fmt"
)

var (
	userCache *cache.Cache
)

func init() {
	// Create a cache with a default expiration time of 5 minutes, and which
	// purges expired items every 10 minutes
	userCache = cache.New(5*time.Minute, 10*time.Minute)
}

func addUserCache(user o2x.User) {
	if user.GetUserID() != nil {
		userCache.Add(fmt.Sprint(user.GetUserID()), user, cache.DefaultExpiration)
	}
}

func getUserCache(id interface{}) (user o2x.User) {
	if c, found := userCache.Get(fmt.Sprint(id)); found {
		user = c.(o2x.User)
		return
	}
	return
}
func removeUserCache(id interface{}) {
	userCache.Delete(fmt.Sprint(id))
}

type MgoUserCfg struct {
	userType reflect.Type

	// password field name
	passwordName string

	// salt field name
	saltName string
}

type MgoUserStore struct {
	session    *mgo.Session
	db         string
	collection string
	userCfg    *MgoUserCfg
}

func DefaultMgoUserCfg() *MgoUserCfg {
	return &MgoUserCfg{
		userType:     o2x.SimpleUserPtrType,
		passwordName: "password",
		saltName:     "salt",
	}
}

func NewUserStore(session *mgo.Session, db, collection string, userCfg *MgoUserCfg) (us *MgoUserStore) {
	if !o2x.IsUserType(userCfg.userType) {
		panic("invalid user type")
	}
	us = &MgoUserStore{
		session:    session,
		db:         db,
		collection: collection,
		userCfg:    userCfg,
	}
	return
}

func (us *MgoUserStore) H(handler func(c *mgo.Collection)) {
	session := us.session.Clone()
	defer session.Close()
	handler(session.DB(us.db).C(us.collection))
	return
}

func (us *MgoUserStore) Save(u o2x.User) (err error) {
	us.H(func(c *mgo.Collection) {
		glog.Infof("insert user:%v", u)
		err = c.Insert(u)
	})

	if err != nil {
		return
	}
	addUserCache(u)

	return
}

func (us *MgoUserStore) Remove(id interface{}) (err error) {
	removeUserCache(id)

	us.H(func(c *mgo.Collection) {
		glog.Infof("remove user:%v", id)
		mgoErr := c.RemoveId(id)
		if mgoErr != nil && mgoErr == mgo.ErrNotFound {
			// try to find using object id
			if sid, ok := id.(string); ok && bson.IsObjectIdHex(sid) {
				bid := bson.ObjectIdHex(sid)
				mgoErr = c.RemoveId(bid)
			}
		}
		if mgoErr != nil && mgoErr == mgo.ErrNotFound {
			err = o2x.ErrNotFound
			return
		}
		err = mgoErr
	})
	return
}

func (us *MgoUserStore) Find(id interface{}) (u o2x.User, err error) {
	if u = getUserCache(id); u != nil {
		return
	}

	us.H(func(c *mgo.Collection) {
		user := o2x.NewUser(us.userCfg.userType)
		mgoErr := c.FindId(id).One(user)
		if mgoErr != nil && mgoErr == mgo.ErrNotFound {
			// try to find using object id
			if sid, ok := id.(string); ok && bson.IsObjectIdHex(sid) {
				bid := bson.ObjectIdHex(sid)
				mgoErr = c.FindId(bid).One(user)
			}
		}

		if mgoErr != nil {
			if mgoErr == mgo.ErrNotFound {
				err = o2x.ErrNotFound
				return
			}
			err = mgoErr
		}

		u = user
	})

	if err != nil {
		return
	}

	if u != nil {
		addUserCache(u)
	}

	return
}

func (us *MgoUserStore) UpdatePwd(id interface{}, password string) (err error) {
	user, err := us.Find(id)
	if err != nil {
		return
	}
	glog.Infof("update user password %v", id)
	user.SetRawPassword(password)

	us.H(func(c *mgo.Collection) {
		bs := bson.M{us.userCfg.passwordName: user.GetPassword(), us.userCfg.saltName: user.GetSalt()}
		bs = bson.M{"$set": bs}
		err = c.UpdateId(user.GetUserID(), bs)
	})

	if err != nil {
		return
	}
	addUserCache(user)
	return
}

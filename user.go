// authors: wangoo
// created: 2018-05-30
// oauth2 user store

package o2m

import (
	"gopkg.in/mgo.v2"
	"github.com/go2s/o2x"
	"reflect"
)

type MgoUserStore struct {
	session    *mgo.Session
	db         string
	collection string
	userType   reflect.Type
}

func NewUserStore(session *mgo.Session, db, collection string, userType reflect.Type) (us *MgoUserStore) {
	if !o2x.IsUserType(userType) {
		panic("invalid user type")
	}
	us = &MgoUserStore{
		session:    session,
		db:         db,
		collection: collection,
		userType:   userType,
	}
	return
}

func (us *MgoUserStore) cHandler(handler func(c *mgo.Collection)) {
	session := us.session.Clone()
	defer session.Close()
	handler(session.DB(us.db).C(us.collection))
	return
}

func (us *MgoUserStore) Save(u o2x.User) (err error) {
	us.cHandler(func(c *mgo.Collection) {
		err = c.Insert(u)
	})
	return
}

func (us *MgoUserStore) Find(id string) (u o2x.User, err error) {
	us.cHandler(func(c *mgo.Collection) {
		user := o2x.NewUser(us.userType)
		mgoErr := c.FindId(id).One(user)
		if mgoErr != nil {
			if mgoErr == mgo.ErrNotFound {
				return
			}
			err = mgoErr
			return
		}

		u = user
	})
	return
}

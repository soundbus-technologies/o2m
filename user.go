// authors: wangoo
// created: 2018-05-30
// oauth2 user store

package mongo

import (
	"gopkg.in/mgo.v2"
	"io"
	"crypto/rand"
	"golang.org/x/crypto/scrypt"
	"log"
	"reflect"
)

const (
	PW_SALT_BYTES = 16
	PW_HASH_BYTES = 32
)

type User struct {
	UserID   string `bson:"_id" json:"user_id"`
	Nickname string `bson:"nickname,omitempty" json:"nickname,omitempty"`
	Password []byte `bson:"password" json:"password"`
	Salt     []byte `bson:"salt" json:"salt"`
}

func (u *User) GetUserID() string {
	return u.UserID
}

func (u *User) SetUserID(userID string) {
	u.UserID = userID
}

func (u *User) GetNickname() string {
	return u.Nickname
}

func (u *User) SetNickname(nickname string) {
	u.Nickname = nickname
}

func (u *User) calcHash(password string) (hash []byte, err error) {
	return scrypt.Key([]byte(password), u.Salt, 1<<14, 8, 1, PW_HASH_BYTES)
}

func (u *User) SetPassword(password string) {
	salt := make([]byte, PW_SALT_BYTES)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		log.Fatal(err)
	}
	u.Salt = salt

	hash, err := u.calcHash(password)
	if err != nil {
		log.Fatal(err)
	}
	u.Password = hash
}

func (u *User) Match(password string) bool {
	hash, err := u.calcHash(password)
	if err != nil {
		log.Fatal(err)
		return false
	}
	return reflect.DeepEqual(hash, u.Password)
}

type UserStore struct {
	session    *mgo.Session
	db         string
	collection string
}

func NewUserStore(session *mgo.Session, db string, collection string) (us *UserStore) {
	us = &UserStore{
		session:    session,
		db:         db,
		collection: collection,
	}
	return
}

func (us *UserStore) cHandler(handler func(c *mgo.Collection)) {
	session := us.session.Clone()
	defer session.Close()
	handler(session.DB(us.db).C(us.collection))
	return
}

func (us *UserStore) Save(u *User) (err error) {
	us.cHandler(func(c *mgo.Collection) {
		err = c.Insert(u)
	})
	return
}

func (us *UserStore) Find(id string) (u *User, err error) {
	us.cHandler(func(c *mgo.Collection) {
		user := &User{}

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

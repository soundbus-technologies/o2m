// authors: wangoo
// created: 2018-06-28
// test user

package o2m

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
	"github.com/go2s/o2x"
	"fmt"
)

const (
	mgoDatabase  = "oauth2"
	mgoUsername  = "oauth2"
	mgoPassword  = "oauth2"
	mgoPoolLimit = 10
)

var mgoAddrs = []string{"127.0.0.1:27017"}

func TestMgoUserStore(t *testing.T) {
	mgoCfg := MongoConfig{
		Addrs:     mgoAddrs,
		Database:  mgoDatabase,
		Username:  mgoUsername,
		Password:  mgoPassword,
		PoolLimit: mgoPoolLimit,
	}

	mgoSession := NewMongoSession(&mgoCfg)

	cfg := DefaultMgoUserCfg()

	us := NewUserStore(mgoSession, mgoDatabase, "user", cfg)

	id := bson.ObjectIdHex("5ae6b2005946fa106132365c")

	fmt.Println("user id:", id.String())

	pass := "123456"
	user, err := us.Find(id)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	if user == nil {
		user = &o2x.SimpleUser{
			UserID: id,
		}
		err = us.Save(user)
		if err != nil {
			assert.Fail(t, err.Error())
		}
	}

	us.UpdatePwd(id, pass)

	updateUser, _ := us.Find(id)

	assert.True(t, updateUser.Match(pass))
	assert.False(t, updateUser.Match("password"))
}

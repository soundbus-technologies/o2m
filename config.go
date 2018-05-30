package mongo

import (
	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
	"fmt"
)


// Config mongodb configuration parameters
type MongoConfig struct {
	Addrs     []string
	Database  string
	Username  string
	Password  string
	PoolLimit int
}

func NewMongoSession(cfg *MongoConfig) *mgo.Session {
	dialInfo := &mgo.DialInfo{
		Addrs:    cfg.Addrs,
		Database: cfg.Database,
		Username: cfg.Username,
		Password: cfg.Password,
	}

	fmt.Printf("dialInfo: %+v\n", dialInfo)

	var err error
	s, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		glog.Infof("Build MongoDB client err : %v", err.Error())
		panic(err)
	}
	s.SetMode(mgo.Monotonic, true)
	s.SetPoolLimit(cfg.PoolLimit)

	fmt.Printf("mongo connected\n")
	return s
}

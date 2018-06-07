package o2m

import (
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

	fmt.Printf("mongo dial info: %+v\n", dialInfo)

	s, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		fmt.Printf("connect mongo error: %v\n", err.Error())
		panic(err)
	}
	s.SetMode(mgo.Monotonic, true)
	s.SetPoolLimit(cfg.PoolLimit)

	fmt.Printf("mongo connected\n")
	return s
}

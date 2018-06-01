# Oauth2 Mongodb Support

mongodb client,token,user storage for oauth2

sample:

```go
package main

import (
	"log"
	"net/http"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3"
)

func main() {
	manager := manage.NewDefaultManager()
    mgoCfg = o2m.MongoConfig{
        Addrs:     []string{"127.0.0.1:27017"},
        Database:  "oauth2",
        Username:  "oauth2",
        Password:  "oauth2",
        PoolLimit: 10,
    }

    mgoSession = o2m.NewMongoSession(&mgoCfg)

    ts := o2m.NewTokenStore(mgoSession, mgoDatabase, "token")

    cs := o2m.NewClientStore(mgoSession, mgoDatabase, "client")

	err = cs.Add(&o2m.Oauth2Client{
		ID:     "000000",
		Secret: "999999",
		Domain: "https://www.baidu.com",
	})
    if err != nil {
        log.Warn("%v\n", err)
    }

    manager.MapClientStorage(cs)
	manager.MustTokenStorage(ts)

    ...

	log.Fatal(http.ListenAndServe(":8080", nil))

}
```
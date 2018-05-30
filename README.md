# MongoDB Storage for OAuth 2.0 Client Information

usage:

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
    mgoCfg = mongo.MongoConfig{
        Addrs:     []string{"127.0.0.1:27017"},
        Database:  "oauth2",
        Username:  "oauth2",
        Password:  "oauth2",
        PoolLimit: 10,
    }

    mgoSession := mongo.NewMongoSession(&mgoCfg)

	store, err := mongo.CreateClientStore(mgoSession, "oauth2", "client")
	if err != nil {
		panic(err)
	}
	err = store.Add(&mongo.Oauth2Client{
		ID:     "000000",
		Secret: "999999",
		Domain: "https://www.baidu.com",
	})
    if err != nil {
        log.Warn("%v\n", err)
    }

    // mongo client store
    manager.MapClientStorage(store)
    // mongo token store
	manager.MustTokenStorage(mongo.CreateTokenStore(mgoSession))

    ...

	log.Fatal(http.ListenAndServe(":8080", nil))

}
```
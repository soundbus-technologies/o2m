// ouath2 client mongo store
// authors: wongoo

package mongo

import (
	"gopkg.in/oauth2.v3"
	"gopkg.in/mgo.v2"
)

const (
	DefaultOuath2ClientDb         = "oauth2"
	DefaultOuath2ClientCollection = "client"
)

// Mongo client store
type MongoClientStore struct {
	db         string
	collection string
	session    *mgo.Session
}

type Oauth2Client struct {
	ID     string `bson:"_id" json:"id"`
	Secret string `bson:"secret" json:"secret"`
	Domain string `bson:"domain" json:"domain"`
	UserID string `bson:"user_id" json:"user_id"`
}

func (c *Oauth2Client) GetID() string {
	return c.ID
}
func (c *Oauth2Client) GetSecret() string {
	return c.Secret
}
func (c *Oauth2Client) GetDomain() string {
	return c.Domain
}
func (c *Oauth2Client) GetUserID() string {
	return c.UserID
}

func NewClientStore(cfg *MongoConfig, db string, collection string) (clientStore *MongoClientStore, err error) {
	return CreateClientStore(NewMongoSession(cfg), db, collection)
}

func CreateClientStore(session *mgo.Session, db string, collection string) (clientStore *MongoClientStore, err error) {
	if session == nil {
		panic("session cannot be nil")
	}
	clientStore = &MongoClientStore{session: session, db: db, collection: collection}
	if clientStore.db == "" {
		clientStore.db = DefaultOuath2ClientDb
	}
	if clientStore.collection == "" {
		clientStore.collection = DefaultOuath2ClientCollection
	}

	return
}

// GetByID according to the ID for the client information
func (cs *MongoClientStore) GetByID(id string) (cli oauth2.ClientInfo, err error) {
	session := cs.session
	defer session.Close()

	c := session.DB(cs.db).C(cs.collection)
	query := c.FindId(id)
	client := &Oauth2Client{}
	err = query.One(client)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Add a client info
func (cs *MongoClientStore) Add(client *Oauth2Client) (err error) {
	session := cs.session
	defer session.Close()

	c := session.DB(cs.db).C(cs.collection)
	return c.Insert(client)
}

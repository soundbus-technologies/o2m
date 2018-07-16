// ouath2 client mongo store
// authors: wongoo

package o2m

import (
	"gopkg.in/oauth2.v3"
	"gopkg.in/mgo.v2"
	"github.com/go2s/o2x"
	"github.com/patrickmn/go-cache"
	"time"
)

const (
	DefaultOauth2ClientDb         = "oauth2"
	DefaultOauth2ClientCollection = "client"
)

var (
	clientCache *cache.Cache
)

func init() {
	// Create a cache with a default expiration time of 5 minutes, and which
	// purges expired items every 10 minutes
	clientCache = cache.New(5*time.Minute, 10*time.Minute)
}

func addClientCache(cli oauth2.ClientInfo) {
	clientCache.Add(cli.GetID(), cli, cache.DefaultExpiration)
}

func getClientCache(id string) (cli oauth2.ClientInfo) {
	if c, found := clientCache.Get(id); found {
		cli = c.(oauth2.ClientInfo)
		return
	}
	return
}

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
	Scope  string `bson:"scope" json:"scope"`
	UserID string `bson:"user_id,omitempty" json:"user_id,omitempty"`
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
func (c *Oauth2Client) GetScope() string {
	return c.Scope
}
func (c *Oauth2Client) GetUserID() string {
	return c.UserID
}

func NewClientStore(session *mgo.Session, db string, collection string) (clientStore *MongoClientStore) {
	if session == nil {
		panic("session cannot be nil")
	}
	clientStore = &MongoClientStore{session: session, db: db, collection: collection}
	if clientStore.db == "" {
		clientStore.db = DefaultOauth2ClientDb
	}
	if clientStore.collection == "" {
		clientStore.collection = DefaultOauth2ClientCollection
	}

	return
}

// GetByID according to the ID for the client information
func (cs *MongoClientStore) GetByID(id string) (cli oauth2.ClientInfo, err error) {
	if cli = getClientCache(id); cli != nil {
		return
	}

	session := cs.session.Clone()
	defer session.Close()

	c := session.DB(cs.db).C(cs.collection)
	query := c.FindId(id)
	client := &Oauth2Client{}
	err = query.One(client)
	if err != nil {
		return nil, err
	}

	addClientCache(client)
	return client, nil
}

// Add a client info
func (cs *MongoClientStore) Set(id string, cli oauth2.ClientInfo) (err error) {
	session := cs.session.Clone()
	defer session.Close()

	c := session.DB(cs.db).C(cs.collection)
	client := &Oauth2Client{
		ID:     cli.GetID(),
		UserID: cli.GetUserID(),
		Domain: cli.GetDomain(),
		Secret: cli.GetSecret(),
	}

	if oauth2Client, ok := cli.(o2x.Oauth2ClientInfo); ok {
		client.Scope = oauth2Client.GetScope()
	}
	addClientCache(client)
	return c.Insert(client)
}

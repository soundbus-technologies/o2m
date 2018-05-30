// authors: wangoo
// created: 2018-05-29
// oauth2 mongodb token storage

package mongo

import (
	"gopkg.in/oauth2.v3"
	"gopkg.in/mgo.v2"
	"time"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2/txn"
	"gopkg.in/oauth2.v3/models"
)

// TokenConfig token configuration parameters
type TokenConfig struct {
	// store db(The default is oauth2)
	DB string
	// store txn collection name(The default is oauth2)
	TxnCName string
	// store token based data collection name(The default is oauth2_basic)
	BasicCName string
	// store access token data collection name(The default is oauth2_access)
	AccessCName string
	// store refresh token data collection name(The default is oauth2_refresh)
	RefreshCName string
}

type basicData struct {
	ID        string       `bson:"_id"`
	Token     models.Token `bson:"Token"`
	ExpiredAt time.Time    `bson:"ExpiredAt"`
}

type tokenData struct {
	ID        string    `bson:"_id"`
	ClientID  string    `bson:"ClientID,omitempty"`
	UserID    string    `bson:"UserID,omitempty"`
	BasicID   string    `bson:"BasicID"`
	ExpiredAt time.Time `bson:"ExpiredAt"`
}

// TokenStore MongoDB storage for OAuth 2.0
type TokenStore struct {
	tcfg    *TokenConfig
	session *mgo.Session
}

// NewDefaultTokenConfig create a default token configuration
func NewDefaultTokenConfig() *TokenConfig {
	return &TokenConfig{
		DB:           "oauth2",
		TxnCName:     "oauth2_txn",
		BasicCName:   "oauth2_basic",
		AccessCName:  "oauth2_access",
		RefreshCName: "oauth2_refresh",
	}
}

// NewTokenStore create a token store instance based on mongodb
func NewTokenStore(cfg *MongoConfig, tcfgs ...*TokenConfig) (store oauth2.TokenStore, err error) {
	var t *TokenConfig = nil
	if len(tcfgs) > 0 {
		t = tcfgs[0]
	}
	return CreateTokenStore(NewMongoSession(cfg), t)
}

// CreateTokenStore create a token store instance based on mongodb
func CreateTokenStore(session *mgo.Session, tcfgs ...*TokenConfig) (store oauth2.TokenStore, err error) {
	ts := &TokenStore{
		session: session,
		tcfg:    NewDefaultTokenConfig(),
	}
	if tcfgs != nil && len(tcfgs) > 0 {
		ts.tcfg = tcfgs[0]
	}

	err = initTokenStore(ts)
	if err != nil {
		return
	}
	store = ts
	return
}

func initTokenStore(ts *TokenStore) (err error) {
	err = ts.c(ts.tcfg.BasicCName).EnsureIndex(mgo.Index{
		Key:         []string{"ExpiredAt"},
		ExpireAfter: time.Second * 1,
	})
	if err != nil {
		return
	}
	err = ts.c(ts.tcfg.AccessCName).EnsureIndex(mgo.Index{
		Key:         []string{"ExpiredAt"},
		ExpireAfter: time.Second * 1,
	})
	if err != nil {
		return
	}
	err = ts.c(ts.tcfg.RefreshCName).EnsureIndex(mgo.Index{
		Key:         []string{"ExpiredAt"},
		ExpireAfter: time.Second * 1,
	})
	return
}

func (ts *TokenStore) c(name string) *mgo.Collection {
	return ts.session.DB(ts.tcfg.DB).C(name)
}

func (ts *TokenStore) cHandler(name string, handler func(c *mgo.Collection)) {
	session := ts.session.Clone()
	defer session.Close()
	handler(session.DB(ts.tcfg.DB).C(name))
	return
}

// Create create and store the new token information
func (ts *TokenStore) Create(info oauth2.TokenInfo) (err error) {
	token := models.Token{
		ClientID:         info.GetClientID(),
		UserID:           info.GetUserID(),
		RedirectURI:      info.GetRedirectURI(),
		Scope:            info.GetScope(),
		Code:             info.GetCode(),
		CodeCreateAt:     info.GetCodeCreateAt(),
		CodeExpiresIn:    info.GetCodeExpiresIn(),
		Access:           info.GetAccess(),
		AccessCreateAt:   info.GetAccessCreateAt(),
		AccessExpiresIn:  info.GetAccessExpiresIn(),
		Refresh:          info.GetRefresh(),
		RefreshCreateAt:  info.GetRefreshCreateAt(),
		RefreshExpiresIn: info.GetRefreshExpiresIn(),
	}

	if code := info.GetCode(); code != "" {
		ts.cHandler(ts.tcfg.BasicCName, func(c *mgo.Collection) {
			err = c.Insert(basicData{
				ID:        code,
				Token:     token,
				ExpiredAt: info.GetCodeCreateAt().Add(info.GetCodeExpiresIn()),
			})
		})
		return
	}

	aexp := info.GetAccessCreateAt().Add(info.GetAccessExpiresIn())
	rexp := aexp
	if info.GetRefresh() != "" && info.GetRefreshExpiresIn() > 0 {
		rexp = info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn())
		if aexp.Second() > rexp.Second() {
			aexp = rexp
		}
	}
	id := bson.NewObjectId().Hex()
	ops := []txn.Op{{
		C:      ts.tcfg.BasicCName,
		Id:     id,
		Assert: txn.DocMissing,
		Insert: basicData{
			Token:     token,
			ExpiredAt: rexp,
		},
	}, {
		C:      ts.tcfg.AccessCName,
		Id:     info.GetAccess(),
		Assert: txn.DocMissing,
		Insert: tokenData{
			BasicID:   id,
			UserID:    info.GetUserID(),
			ClientID:  info.GetClientID(),
			ExpiredAt: aexp,
		},
	}}
	if refresh := info.GetRefresh(); refresh != "" {
		ops = append(ops, txn.Op{
			C:      ts.tcfg.RefreshCName,
			Id:     refresh,
			Assert: txn.DocMissing,
			Insert: tokenData{
				BasicID:   id,
				ExpiredAt: rexp,
			},
		})
	}
	ts.cHandler(ts.tcfg.TxnCName, func(c *mgo.Collection) {
		runner := txn.NewRunner(c)
		err = runner.Run(ops, "", nil)
	})
	return
}

// RemoveByCode use the authorization code to delete the token information
func (ts *TokenStore) RemoveByCode(code string) (err error) {
	ts.cHandler(ts.tcfg.BasicCName, func(c *mgo.Collection) {
		verr := c.RemoveId(code)
		if verr != nil {
			if verr == mgo.ErrNotFound {
				return
			}
			err = verr
		}
	})
	return
}

// RemoveByAccess use the access token to delete the token information
func (ts *TokenStore) RemoveByAccess(access string) (err error) {
	ts.cHandler(ts.tcfg.AccessCName, func(c *mgo.Collection) {
		verr := c.RemoveId(access)
		if verr != nil {
			if verr == mgo.ErrNotFound {
				return
			}
			err = verr
		}
	})
	return
}

// RemoveByRefresh use the refresh token to delete the token information
func (ts *TokenStore) RemoveByRefresh(refresh string) (err error) {
	ts.cHandler(ts.tcfg.RefreshCName, func(c *mgo.Collection) {
		verr := c.RemoveId(refresh)
		if verr != nil {
			if verr == mgo.ErrNotFound {
				return
			}
			err = verr
		}
	})
	return
}

func (ts *TokenStore) getData(basicID string) (ti oauth2.TokenInfo, err error) {
	ts.cHandler(ts.tcfg.BasicCName, func(c *mgo.Collection) {
		var bd basicData
		verr := c.FindId(basicID).One(&bd)
		if verr != nil {
			if verr == mgo.ErrNotFound {
				return
			}
			err = verr
			return
		}

		ti = &bd.Token
	})
	return
}

func (ts *TokenStore) getBasicID(cname, token string) (basicID string, err error) {
	ts.cHandler(cname, func(c *mgo.Collection) {
		var td tokenData
		verr := c.FindId(token).One(&td)
		if verr != nil {
			if verr == mgo.ErrNotFound {
				return
			}
			err = verr
			return
		}
		basicID = td.BasicID
	})
	return
}

// GetByCode use the authorization code for token information data
func (ts *TokenStore) GetByCode(code string) (ti oauth2.TokenInfo, err error) {
	ti, err = ts.getData(code)
	return
}

// GetByAccess use the access token for token information data
func (ts *TokenStore) GetByAccess(access string) (ti oauth2.TokenInfo, err error) {
	basicID, err := ts.getBasicID(ts.tcfg.AccessCName, access)
	if err != nil && basicID == "" {
		return
	}
	ti, err = ts.getData(basicID)
	return
}

// GetByRefresh use the refresh token for token information data
func (ts *TokenStore) GetByRefresh(refresh string) (ti oauth2.TokenInfo, err error) {
	basicID, err := ts.getBasicID(ts.tcfg.RefreshCName, refresh)
	if err != nil && basicID == "" {
		return
	}
	ti, err = ts.getData(basicID)
	return
}

// getBasicIDByAccount get the basic id by userID and clientID
func (ts *TokenStore) getBasicIDByAccount(cname, userID string, clientID string) (basicID string, err error) {
	ts.cHandler(cname, func(c *mgo.Collection) {
		var td tokenData
		verr := c.Find(bson.M{"UserID": userID, "ClientID": clientID}).One(&td)
		if verr != nil {
			if verr == mgo.ErrNotFound {
				return
			}
			err = verr
			return
		}
		basicID = td.BasicID
	})
	return
}

// GetByAccount get the exists token info by userID and clientID
func (ts *TokenStore) GetByAccount(userID string, clientID string) (ti oauth2.TokenInfo, err error) {
	basicID, err := ts.getBasicIDByAccount(ts.tcfg.RefreshCName, userID, clientID)
	if err != nil && basicID == "" {
		return
	}
	ti, err = ts.getData(basicID)
	return
}

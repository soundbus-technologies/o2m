// authors: wangoo
// created: 2018-05-30
// test user

package mongo

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
	"encoding/json"
)

func TestUser(t *testing.T) {
	u := &User{}
	u.SetUserID("123")
	u.SetNickname("wongoo")

	assert.Equal(t, "123", u.GetUserID())
	assert.Equal(t, "wongoo", u.GetNickname())

	password := "my_password"
	u.SetPassword(password)

	assert.True(t, u.Match(password), "password should be match")

	json, err := json.Marshal(u)
	assert.Nil(t, err, err)
	fmt.Println(string(json))
}

/**------------------------------------------------------------**
 * @filename auth/basic.go
 * @author   jinycoo
 * @version  1.0.0
 * @date     2019-07-24 16:28
 * @desc     go.jd100.com - auth - basic auth
 **------------------------------------------------------------**/
package auth

import (
	"encoding/base64"
	"net/http"
	"strconv"

	"go.jd100.com/medusa/net/http/jiny"
)

func Basic(accounts map[string]string, realm string) jiny.HandlerFn {
	if realm == "" {
		realm = "Authorization Required"
	}
	realm = "Basic realm=" + strconv.Quote(realm)
	pairs := processAccounts(accounts)
	return func(c *jiny.Context) {
		account, found := pairs.searchCredential(c.Request.Header.Get(ReqAuthKey))
		if !found {
			c.Header(ResAuthKey, realm)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set(Account, account)
	}
}

func (a authPairs) searchCredential(authValue string) (string, bool) {
	if authValue == "" {
		return "", false
	}
	for _, pair := range a {
		if pair.value == authValue {
			return pair.user, true
		}
	}
	return "", false
}

func processAccounts(accounts Accounts) authPairs {
	pairs := make(authPairs, 0, len(accounts))
	for user, password := range accounts {
		value := authorizationHeader(user, password)
		pairs = append(pairs, authPair{
			value: value,
			user:  user,
		})
	}
	return pairs
}

func authorizationHeader(user, password string) string {
	base := user + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(base))
}

package auth

import "github.com/gorilla/securecookie"

type Session struct {
	Secure *securecookie.SecureCookie
	Key    string
}

func NewSecureCookie(hashKey, blockKey string) *Session {
	var bk []byte
	if blockKey != "" {
		bk = []byte(blockKey)
	}
	sc := securecookie.New([]byte(hashKey), bk)
	return &Session{Secure: sc, Key: "session"}
}


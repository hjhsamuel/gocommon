package gin_plugin

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Context struct {
	*gin.Context
	User      *UserAuth
	BasicAuth *BasicAuth
}

type UserAuth struct {
	ID   int
	Name string
	jwt.RegisteredClaims
}

type BasicAuth struct {
	Name     string
	Password string
}

const (
	authKey      = "Auth"
	basicAuthKey = "BasicAuth"
)

var (
	cookieName string
	tokenSalt  string
	ctxLock    sync.RWMutex

	maxPrintBodySize = 1024 * 4
)

func SetTokenSalt(salt string) {
	ctxLock.Lock()
	defer ctxLock.Unlock()

	tokenSalt = salt
}

func GetTokenSalt() string {
	ctxLock.RLock()
	defer ctxLock.RUnlock()

	return tokenSalt
}

func SetCookieName(name string) {
	ctxLock.Lock()
	defer ctxLock.Unlock()

	cookieName = name
}

func GetCookieName() string {
	ctxLock.RLock()
	defer ctxLock.RUnlock()

	return cookieName
}

func SetMaxPrintBodySize(n int) {
	maxPrintBodySize = n
}

func GetMaxPrintBodySize() int {
	return maxPrintBodySize
}

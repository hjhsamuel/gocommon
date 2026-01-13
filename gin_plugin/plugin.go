package gin_plugin

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

func Auth(ctx *gin.Context) {
	var info string
	content := ctx.Request.Header.Get("Authorization")
	if content == "" {
		name := GetCookieName()
		if v, err := ctx.Cookie(name); err == nil {
			info = v
		}
	} else {
		parts := strings.Fields(content)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			info = parts[1]
		}
	}

	if info != "" {
		claim := &UserAuth{}
		token, err := jwt.ParseWithClaims(info, claim, func(token *jwt.Token) (any, error) {
			salt := GetTokenSalt()
			return []byte(salt), nil
		})
		if err == nil && token.Valid {
			ctx.Set(authKey, claim)
		}
	}
	ctx.Next()
}

func BasicAuthMiddleware(ctx *gin.Context) {
	username, password, ok := ctx.Request.BasicAuth()
	if ok {
		ctx.Set(basicAuthKey, &BasicAuth{
			Name:     username,
			Password: password,
		})
	}
	ctx.Next()
}

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	_, _ = w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func ReqLogger(ctx *gin.Context) {
	startTime := time.Now()

	fields := logrus.Fields{"Method": ctx.Request.Method, "ClientIP": ctx.ClientIP()}

	var (
		bw      *bodyWriter
		maxBody = GetMaxPrintBodySize()
	)
	if logrus.GetLevel() >= logrus.DebugLevel {
		// query
		queryParams := ctx.Request.URL.Query()
		if len(queryParams) != 0 {
			val := new(strings.Builder)
			val.WriteString("{")
			for k, v := range queryParams {
				val.WriteString(fmt.Sprintf("%s: %v,", k, v))
			}
			val.WriteString("}")
			fields["QueryParams"] = val.String()
		}
		// body
		if body, err := ctx.GetRawData(); err == nil && len(body) > 0 {
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			if len(body) <= maxBody {
				fields["Body"] = string(body)
			}
		}
		// replace response writer
		bw = &bodyWriter{body: &bytes.Buffer{}, ResponseWriter: ctx.Writer}
		ctx.Writer = bw
	}

	ctx.Next()

	fields["Code"] = ctx.Writer.Status()

	if logrus.GetLevel() >= logrus.DebugLevel {
		// response body
		if bw != nil {
			if maxBody == 0 || bw.body.Len() <= maxBody {
				fields["Response"] = bw.body.String()
			}
			bw.body.Reset()
		}
	}

	fields["Cost"] = time.Since(startTime)
	logrus.WithFields(fields).Info(ctx.Request.URL.Path)
}

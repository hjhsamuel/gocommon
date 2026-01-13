package gin_plugin

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	ErrAuth        = Error(401, "authorization failed") // 认证失败
	ErrRedirect    = Error(302, "redirect")             // 重定向
	ErrRawResponse = Error(200, "return raw response")  // 使用原始响应
)

type Handler[T1, T2 any] func(ctx *Context, req T1) (T2, error)

func GetHandler[T1, T2 any](h Handler[*T1, T2], withAuth bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var newErr error
				switch err.(type) {
				case error:
					newErr = err.(error)
				default:
					newErr = fmt.Errorf("%v", err)
				}
				logrus.WithError(newErr).Error("panic")
				ctx.JSON(
					http.StatusOK,
					&Response{
						Code:    http.StatusInternalServerError,
						Message: newErr.Error(),
					},
				)
			}
		}()

		var (
			req = new(T1)
			err error
		)
		// bug exists: DO NOT USE `required` tag
		// get uri params
		_ = ctx.BindUri(req)
		// get query params
		_ = ctx.BindQuery(req)
		// get request body
		if ctx.Request.Method != http.MethodGet {
			// 兼容文件上传
			if ctx.ContentType() == "multipart/form-data" {
				err = ctx.ShouldBind(req)
			} else {
				err = ctx.ShouldBindJSON(req)
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			} else {
				ctx.JSON(
					http.StatusOK,
					&Response{
						Code:    http.StatusInternalServerError,
						Message: err.Error(),
					},
				)
				return
			}
		}

		c := &Context{Context: ctx}
		if v, ok := ctx.Get(authKey); ok && v != nil {
			if claim, fok := v.(*UserAuth); fok {
				c.User = claim
			}
		}
		if v, ok := ctx.Get(basicAuthKey); ok && v != nil {
			if claim, fok := v.(*BasicAuth); fok {
				c.BasicAuth = claim
			}
		}
		if withAuth && c.User == nil && c.BasicAuth == nil {
			ctx.JSON(
				http.StatusOK,
				&Response{
					Code:    ErrAuth.Code(),
					Message: ErrAuth.Error(),
				},
			)
			return
		}

		out, err := h(c, req)
		if err != nil {
			rsp := &Response{}
			switch hErr := err.(type) {
			case *EZErr:
				if errors.Is(hErr, ErrRedirect) {
					if path, ok := any(out).(string); ok {
						ctx.Redirect(http.StatusFound, path)
						return
					}
				}
				if errors.Is(hErr, ErrRawResponse) {
					return
				}
				rsp.Code = hErr.Code()
				rsp.Message = hErr.Error()
			default:
				rsp.Code = http.StatusInternalServerError
				rsp.Message = err.Error()
			}
			ctx.JSON(http.StatusOK, rsp)
		} else {
			ctx.JSON(
				http.StatusOK,
				&Response{
					Code: http.StatusOK,
					Data: out,
				},
			)
		}
	}
}

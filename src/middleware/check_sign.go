package middleware

import (
	"bytes"
	"io"

	"github.com/assimon/luuu/config"
	"github.com/assimon/luuu/util/constant"
	"github.com/assimon/luuu/util/json"
	"github.com/assimon/luuu/util/sign"
	"github.com/labstack/echo/v4"
)

func CheckApiSign() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			params, err := io.ReadAll(ctx.Request().Body)
			if err != nil {
				return constant.SignatureErr
			}
			m := make(map[string]interface{})
			err = json.Cjson.Unmarshal(params, &m)
			if err != nil {
				return err
			}
			signature, ok := m["signature"]
			if !ok {
				return constant.SignatureErr
			}
			checkSignature, err := sign.Get(m, config.GetApiAuthToken())
			if err != nil {
				return constant.SignatureErr
			}
			if checkSignature != signature {
				return constant.SignatureErr
			}
			ctx.Request().Body = io.NopCloser(bytes.NewBuffer(params))
			return next(ctx)
		}
	}
}

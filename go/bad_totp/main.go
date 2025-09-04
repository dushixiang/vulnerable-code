package main

import (
	"crypto/hmac"
	"crypto/sha1"
	_ "embed"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	secret   = "*****" // 隐藏密钥
	timeStep = 30      // 时间步长，单位为秒
	digits   = 6       // totp长度
)

func generateTOTP(secret string, t time.Time) string {
	// 解码密钥
	key, _ := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	// 每30秒生成一个新的TOTP码
	n := t.Unix() / int64(timeStep)
	// 将时间戳转换为8个字节的大端字节序
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(n))

	// 计算HMAC
	h := hmac.New(sha1.New, key)
	h.Write(msg)
	hmacHash := h.Sum(nil)

	// 计算TOTP令牌
	offset := hmacHash[len(hmacHash)-1] & 0x0f
	truncatedHash := binary.BigEndian.Uint32(hmacHash[offset:offset+4]) & 0x7fffffff
	totpCode := fmt.Sprintf("%0*d", digits, truncatedHash%uint32(10^digits))
	// 将结果转换为6位的字符串
	return totpCode
}

func validateTOTP(secret, passcode string, t time.Time) bool {
	expectedCode := generateTOTP(secret, t)
	// 验证提供的TOTP码和预期的TOTP码是否匹配
	return passcode == expectedCode
}

//go:embed index.html
var indexHtml string

func main() {
	e := echo.New()
	e.Use(middleware.Gzip())
	e.Debug = true

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, indexHtml)
	})

	e.POST("/flag", func(c echo.Context) error {
		passcode := c.FormValue("passcode")
		if !validateTOTP(secret, passcode, time.Now()) {
			return c.String(400, `您输入的密码不正确`)
		}
		return c.String(200, os.Getenv("flag"))
	})

	e.Logger.Fatal(e.Start(":80"))
}

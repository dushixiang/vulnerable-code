package main

import (
	"encoding/base64"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var notBase64Std = "/+9876543210zyxwvutsrqponmlkjihgfedcbaZYXWVUTSRQPONMLKJIHGFEDCBA"
var notStdEncoding = base64.NewEncoding(notBase64Std)

func main() {
	e := echo.New()
	e.Debug = true
	e.Use(middleware.Gzip())

	e.GET("/", func(c echo.Context) error {
		flag := os.Getenv("flag")
		encoded := notStdEncoding.EncodeToString([]byte(flag))
		return c.String(http.StatusOK, encoded)
	})

	e.Logger.Fatal(e.Start(":80"))
}

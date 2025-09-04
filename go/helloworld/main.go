package main

import (
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.Debug = true
	e.Use(middleware.Gzip())

	e.GET("/", func(c echo.Context) error {
		flag := os.Getenv("flag")

		cookie := new(http.Cookie)
		cookie.Name = "flag"
		cookie.Value = flag
		cookie.Expires = time.Now().Add(24 * time.Hour)
		c.SetCookie(cookie)

		return c.HTML(http.StatusOK, `Hello World!`)
	})

	e.Logger.Fatal(e.Start(":80"))
}

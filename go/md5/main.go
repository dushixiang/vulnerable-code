package main

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

const (
	username = "cyberpoc"
	password = "a3d0715cb0cfb6113e9eb87a020e94ea"
)

func hash(input string) string {
	data := []byte(input)
	sum := md5.Sum(data)
	md5str := fmt.Sprintf("%x", sum)
	return md5str
}

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `<a href='/admin'>Admin Dashboard</a>`)
	})

	e.GET("/admin", func(c echo.Context) error {
		u := c.Request().Header.Get("username")
		p := c.Request().Header.Get("password")
		if u != username || hash(p) != password {
			return echo.ErrUnauthorized
		}
		return c.String(200, os.Getenv("flag"))
	})

	e.Logger.Fatal(e.Start(":80"))
}

package main

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var blockList = []string{"127.0.0.1", "localhost"}

func cyberpocSSRF(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		target := c.QueryParam("target")
		_url, err := url.Parse(strings.TrimSpace(target))
		if err != nil {
			return err
		}
		hostname := _url.Hostname()
		for _, block := range blockList {
			if hostname == block {
				return c.String(http.StatusBadRequest, "Not allowed!")
			}
		}

		return next(c)
	}
}

//go:embed index.html
var indexHtml string

func main() {

	go func() {
		debug := echo.New()
		debug.GET("/flag", func(c echo.Context) error {
			html := fmt.Sprintf(`FLAG: %s`, os.Getenv("flag"))
			return c.HTML(http.StatusOK, html)
		})
		debug.Logger.Fatal(debug.Start("127.0.0.1:8000"))
	}()

	e := echo.New()
	e.Debug = true
	e.Use(middleware.Gzip())

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, indexHtml)
	})
	e.GET("/get-file", func(c echo.Context) error {
		target := c.QueryParam("target")
		resp, err := http.Get(strings.TrimSpace(target))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		_, err = c.Response().Write(bytes)
		return err
	}, cyberpocSSRF)
	e.Logger.Fatal(e.Start(":80"))
}

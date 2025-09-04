package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func cyberpocSSRF(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		target := c.QueryParam("target")
		_url, err := url.Parse(strings.TrimSpace(target))
		if err != nil {
			return err
		}

		ips, err := net.LookupIP(_url.Hostname())
		if err != nil {
			return err
		}
		for _, ip := range ips {
			ipAddr := ip.String()
			if ipAddr == "82.157.160.150" {
				return next(c)
			}
		}
		return c.String(http.StatusBadRequest, "Not allowed!")
	}
}

const indexHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>SSRF | CyberPoC</title>
</head>
<body>
<H1>CyberPoC SSRF</H1>

<form action="/get-file" method="get">
    <input type="text" name="target" value="https://f.typesafe.cn/cyberpoc/hi.txt">
    <button type="submit">Get</button>
</form>
</body>
</html>`

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
		// optimize when you pay
		time.Sleep(time.Second * 3)
		// get file
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

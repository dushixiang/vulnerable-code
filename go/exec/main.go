package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type CyberWriter struct {
	resp *echo.Response
}

func (r CyberWriter) Write(p []byte) (n int, err error) {
	n, err = r.resp.Write(p)
	r.resp.Flush()
	return
}

//go:embed index.html
var indexHtml string

func main() {

	e := echo.New()
	e.Use(middleware.Gzip())
	e.Use(middleware.CORS())
	e.Debug = true

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, indexHtml)
	})

	e.GET("/ping", func(c echo.Context) error {
		ip := c.QueryParam("ip")
		if ip == "" {
			return c.String(400, "ip is required")
		}

		command := fmt.Sprintf(`ping %s -c 4`, ip)

		cmd := exec.Command("sh", "-c", command)

		log.Println("run command", cmd.String())

		c.Response().Header().Set(echo.HeaderContentType, `text/event-stream`)
		c.Response().WriteHeader(http.StatusOK)

		writer := CyberWriter{
			resp: c.Response(),
		}
		cmd.Stdout = writer
		cmd.Stderr = writer
		err := cmd.Run()
		if err != nil {
			return err
		}
		return nil
	})

	e.Logger.Fatal(e.Start(":80"))
}

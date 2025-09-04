package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4/middleware"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.Debug = true
	e.IPExtractor = echo.ExtractIPFromRealIPHeader()
	e.Use(middleware.Gzip())

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `<H1>CyberPoC SSRF</H1><br/>
<a href='/admin'>Admin Dashboard</a>`)
	})

	e.GET("/admin", func(c echo.Context) error {
		_, _ = fmt.Fprintln(c.Response(), `<p>Verify that your an administrator, wait...</p>`)

		clientIp := c.RealIP()
		cookie, _ := c.Request().Cookie("role")
		if cookie == nil {
			return c.HTML(http.StatusForbidden, `<p>You are not administrator</p>`)
		}
		role := cookie.Value

		var passed = false
		if strings.EqualFold(role, "admin") {
			for _, ip := range []string{"127.0.0.1", "localhost"} {
				if ip == clientIp {
					passed = true
					break
				}
			}
		}

		log.Printf("clientIp: %s, role: %s, passed: %v\n", clientIp, role, passed)

		if passed {
			html := fmt.Sprintf(`FLAG: %s`, os.Getenv("flag"))
			_, _ = fmt.Fprintln(c.Response(), html)
			return nil
		} else {
			_, _ = fmt.Fprintln(c.Response(), `<p>You are not administrator.</p>`)
			return nil
		}
	})

	e.Logger.Fatal(e.Start(":80"))
}

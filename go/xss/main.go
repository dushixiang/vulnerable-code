package main

import (
	"html/template"
	"net/http"
	"os"
	"time"

	"xss/cyberpoc"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Gzip())
	e.Use(middleware.Logger())
	e.Debug = true

	e.Renderer = &cyberpoc.TemplateRenderer{
		Templates: template.Must(template.ParseGlob("*.html")),
	}

	letterRepo := new(cyberpoc.LetterRepo)

	e.GET("/", func(c echo.Context) error {

		return c.Render(http.StatusOK, `index.html`, map[string]interface{}{})
	})

	e.POST("/letter", func(c echo.Context) error {
		mail := c.FormValue("mail")
		title := c.FormValue("title")
		content := c.FormValue("content")
		letter := cyberpoc.Letter{
			Mail:    mail,
			Title:   title,
			Content: template.HTML(content),
		}
		letterRepo.Save(letter)
		return c.JSON(200, map[string]string{})
	})

	e.GET("/letters", func(c echo.Context) error {
		letters := letterRepo.List()

		return c.Render(http.StatusOK, "letter.html", map[string]interface{}{
			"letters": letters,
		})
	})

	ticker := time.NewTicker(time.Second * 5)
	// 定时器、访问地址、cookies
	go cyberpoc.RunChrome(ticker, "http://localhost/letters", map[string]string{
		"token": os.Getenv("flag"),
	})

	e.Logger.Fatal(e.Start(":80"))
}

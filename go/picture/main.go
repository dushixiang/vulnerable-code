package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const indexHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Image Backup | CyberPoC</title>
</head>
<body>
<H1>CyberPoC Image Backup</H1>

<form action="/picture" method="post" enctype="application/x-www-form-urlencoded">
	<label for="url"><b>url</b></label>
    <input type="text" placeholder="Enter url" name="url" required>

    <label for="image"><b>image</b></label>
    <input type="img" placeholder="Enter image name" name="img" required>
    <button type="submit">Login</button>
</form>
</body>
</html>`

func writeFile(filename string, content []byte) error {
	_, err := os.Stat(filename)
	if err == nil {
		_ = os.Remove(filename)
	}

	return os.WriteFile(filename, content, os.ModePerm)
}

func main() {
	e := echo.New()
	e.Debug = true
	e.Use(middleware.Gzip())

	e.GET("/", func(c echo.Context) error {
		return c.HTML(200, indexHtml)
	})

	e.POST("/picture", func(c echo.Context) error {
		url := c.FormValue("url")
		if url == "" {
			url = "http://localhost"
		}
		img := c.FormValue("img")
		if img == "" {
			img = "profile.jpg"
		}

		if !strings.Contains(url, "http://localhost") {
			return c.String(403, `Forbidden 1`)
		}

		ext := path.Ext(img)
		if ext != ".jpg" && ext != ".png" {
			return c.String(403, `Forbidden 2`)
		}

		var args = []string{
			"curl", "--proto", "-file", strconv.Quote(url), "-X",
			"GET", "-F", strconv.Quote(fmt.Sprintf(`image=%s`, img)), ">",
			"backup_profile.jpg",
		}

		cmd := exec.Command("sh", "-c", strings.Join(args, " "))

		if err := cmd.Run(); err != nil {
			return c.String(500, `Internal Server Error: `+err.Error())
		}

		return c.String(200, "thanks for testing our image backup service!")
	})

	flag := os.Getenv("flag")
	if err := writeFile("flag", []byte(flag)); err != nil {
		e.Logger.Fatal(err)
	}

	e.Logger.Fatal(e.Start(":80"))
}

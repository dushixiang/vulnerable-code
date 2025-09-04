package main

import (
	_ "embed"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed index.html
var indexHtml string

func main() {

	const uploadDir = "files"

	e := echo.New()
	e.Debug = true
	e.Use(middleware.Gzip())

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, indexHtml)
	})

	e.GET("/download", func(c echo.Context) error {
		filename := c.QueryParam("filename")
		return c.File(path.Join(uploadDir, filename))
	})

	e.POST("/upload", func(c echo.Context) error {
		file, err := c.FormFile("file")
		if err != nil {
			return err
		}
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		fullName := path.Join(uploadDir, file.Filename)
		// Destination
		dst, err := os.Create(fullName)
		if err != nil {
			return err
		}
		defer dst.Close()

		// Copy
		if _, err = io.Copy(dst, src); err != nil {
			return err
		}

		return c.JSON(200, map[string]string{
			"filename": file.Filename,
		})
	})

	_, err := os.Stat(uploadDir)
	if err != nil {
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			e.Logger.Fatal(err)
		}
	}

	_ = os.Remove("flag")
	flag := os.Getenv("flag")
	if err := os.WriteFile("flag", []byte(flag), os.ModeType); err != nil {
		e.Logger.Fatal(err)
	}

	e.Logger.Fatal(e.Start(":80"))
}

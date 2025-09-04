package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func writeFile(filename string, content []byte) error {
	_, err := os.Stat(filename)
	if err == nil {
		_ = os.Remove(filename)
	}

	return os.WriteFile(filename, content, os.ModePerm)
}

func main() {

	const uploadDir = "files"
	const indexHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Upload | CyberPoC</title>
</head>
<body>
<H1>CyberPoC Upload</H1>

<form action="/upload" method="post" enctype="multipart/form-data">
    <input type="file" name="file">
    <button type="submit">Upload</button>
</form>
</body>
</html>`

	e := echo.New()
	e.Use(middleware.Gzip())
	e.Debug = true

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, indexHtml)
	})

	e.GET("/download", func(c echo.Context) error {
		filename := c.QueryParam("filename")
		filename = strings.ReplaceAll(filename, `/`, ``)
		filename = strings.ReplaceAll(filename, `\`, ``)
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
		filename := c.QueryParam("filename")
		if filename == "" {
			filename = file.Filename
		}

		fullName := path.Join(uploadDir, filename)
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

		// Detect virus
		cmd := exec.Command("sh", "detect_virus.sh", fullName)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("err: %s, stderr: %s", err.Error(), stderr.String()))
		}

		return c.HTML(http.StatusOK, fmt.Sprintf(`<h2>Uploaded successfully!</h2><br/>
<a href='/'>Home</a><br/>
<a href='/download?filename=%s'>Download</a><br/>
`, filename))
	})

	_, err := os.Stat(uploadDir)
	if err != nil {
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			e.Logger.Fatal(err)
		}
	}

	flag := os.Getenv("flag")
	if err := writeFile("flag", []byte(flag)); err != nil {
		e.Logger.Fatal(err)
	}
	if err := writeFile("detect_virus.sh", []byte("echo passed")); err != nil {
		e.Logger.Fatal(err)
	}

	e.Logger.Fatal(e.Start(":80"))
}

package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const uploadDir = "files"
const sessionDir = "sessions"

type User struct {
	Username string
	IsAdmin  bool
}

func cyberpocAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		u, err := getUser(c)
		if err != nil {
			return err
		}

		if !u.IsAdmin {
			return c.HTML(http.StatusForbidden, `<p>You are not administrator.</p>`)
		}
		return next(c)
	}
}

func cyberpocAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := getUser(c)
		if err != nil {
			return err
		}
		return next(c)
	}
}

func getUser(c echo.Context) (u User, err error) {
	cookie, err := c.Cookie("session")
	if err != nil {
		return
	}
	if cookie.Value == "" {
		err = errors.New("not Login")
		return
	}
	return getUserBySession(cookie.Value)
}

func getUserBySession(session string) (u User, err error) {
	var f = path.Join(sessionDir, session)
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return
	}
	decoder := gob.NewDecoder(bytes.NewReader(b))
	err = decoder.Decode(&u)
	return
}

const loginHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>GOB | CyberPoC</title>
</head>
<body>
<H1>CyberPoC GOB</H1>

<form action="/login" method="post" enctype="application/x-www-form-urlencoded">
	<label for="uname"><b>Username</b></label>
    <input type="text" placeholder="Enter Username" name="username" required>

    <label for="psw"><b>Password</b></label>
    <input type="password" placeholder="Enter Password" name="password" required>
    <button type="submit">Login</button>
</form>
</body>
</html>`

func main() {

	e := echo.New()
	e.Debug = true
	e.Use(middleware.Gzip())

	e.GET("/", func(c echo.Context) error {
		u, err := getUser(c)
		if err != nil {
			return c.HTML(http.StatusOK, loginHtml)
		}

		return c.HTML(http.StatusOK, `<H1>CyberPoC GOB: `+u.Username+`</H1><br/>
<form action="/upload" method="post" enctype="multipart/form-data">
    <input type="file" name="file">
    <button type="submit">Upload</button>
</form>
</body>
<a href='/admin'>Admin Dashboard</a>`)
	})

	e.GET("/admin", func(c echo.Context) error {
		return c.String(http.StatusOK, os.Getenv("flag"))
	}, cyberpocAdmin)

	e.POST("/login", func(c echo.Context) error {
		username := c.FormValue("username")
		var u = User{
			Username: username,
			IsAdmin:  false,
		}

		session := strconv.FormatInt(time.Now().UnixMicro(), 10)
		var f = path.Join(sessionDir, session)

		file, err := os.Create(f)
		if err != nil {
			return err
		}

		encoder := gob.NewEncoder(file)
		if err := encoder.Encode(u); err != nil {
			return err
		}

		cookie := http.Cookie{
			Name:    "session",
			Value:   session,
			Path:    "/",
			Domain:  "",
			Expires: time.Now().Add(time.Hour),
		}

		time.AfterFunc(time.Hour, func() {
			_ = os.RemoveAll(f)
		})

		c.SetCookie(&cookie)
		return c.Redirect(302, "/")
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

		return c.HTML(http.StatusOK, fmt.Sprintf(`<h2>Uploaded successfully!</h2><br/>
<a href='/'>Home</a><br/>
<a href='/download?filename=%s'>Download</a><br/>
`, filename))
	}, cyberpocAuth)

	_, err := os.Stat(uploadDir)
	if err != nil {
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			e.Logger.Fatal(err)
		}
	}

	_, err = os.Stat(sessionDir)
	if err != nil {
		if err := os.MkdirAll(sessionDir, os.ModePerm); err != nil {
			e.Logger.Fatal(err)
		}
	}

	e.Logger.Fatal(e.Start(":80"))
}

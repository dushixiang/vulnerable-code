package main

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const sessionDir = "sessions"

const indexHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Rank | CyberPoC</title>
</head>
<body>
<H1>CyberPoC Rank</H1> 
<a href="/login"> Login </a>
<table>
    <tr>
        <th>Index</th>
        <th>Username</th>
        <th>Score</th>
        <th>Detail</th>
    </tr>
    {{ range .Items}}
        <tr>
            <td>{{ .Index }}</td>
            <td>{{ .Username }}</td>
            <td>{{ .Score }}</td>
            <td><a href="/users/{{ .Username }}?csrf={{ $.CSRF }}">go</td>
        </tr>
    {{ end}}
</table>
</body>
</html>`

const loginHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Rank | CyberPoC</title>
</head>
<body>
<H1>CyberPoC Rank</H1>

<form action="/login" method="post" enctype="application/x-www-form-urlencoded">
	<label for="uname"><b>Username</b></label>
    <input type="text" placeholder="Enter Username" name="username" required>

    <label for="psw"><b>Password</b></label>
    <input type="password" placeholder="Enter Password" name="password" required>
    <button type="submit">Login</button>
</form>
</body>
</html>`

type Data struct {
	Items []User
	CSRF  string
}

type User struct {
	Index    int
	Username string
	Password string
	Score    int
}

var (
	ranks  []User
	locker sync.Mutex
)

func md5X(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func getUsers() []User {
	locker.Lock()
	defer locker.Unlock()

	if len(ranks) > 0 {
		return ranks
	}

	for i := 0; i < 10; i++ {
		ranks = append(ranks, User{
			Index:    i + 1,
			Username: fmt.Sprintf(`user_%02d`, i+1),
			Password: md5X(strconv.Itoa(i + 1)),
			Score:    10000 - (i+1)*100,
		})
	}

	return ranks
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
		err = errors.New("not login")
		return
	}
	decoder := gob.NewDecoder(bytes.NewReader(b))
	err = decoder.Decode(&u)
	return
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

func main() {
	e := echo.New()
	e.Debug = true
	e.Use(middleware.Gzip())

	e.GET("/", func(c echo.Context) error {
		t, err := template.New("index").Parse(indexHtml)
		if err != nil {
			return err
		}

		data := Data{
			Items: getUsers(),
			CSRF:  md5X(strconv.FormatInt(time.Now().UnixMicro(), 10)),
		}
		return t.Execute(c.Response(), data)
	})

	e.GET("/login", func(c echo.Context) error {
		return c.HTML(200, loginHtml)
	})

	e.GET("/users/:username", func(c echo.Context) error {
		username := c.Param("username")
		csrf := c.QueryParam("csrf")

		var temp = fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Rank | CyberPoC</title>
</head>
<body>
<H1>CyberPoC Rank</H1>
<br/>
<p>You are looking for: %s</p>

<br/>
<p><strong> Username: {{ .Username }} </strong></p>
<p><strong>    Score: {{ .Score }} </strong></p>
<div style="display: none;">%s<div>
</body>
</html>`, username, csrf)

		t, err := template.New("user").Parse(temp)

		if err != nil {
			return err
		}

		var u User
		users := getUsers()
		for _, user := range users {
			if user.Username == username {
				u = user
				break
			}
		}

		return t.Execute(c.Response(), &u)
	})

	e.GET("/flag", func(c echo.Context) error {
		return c.String(http.StatusOK, os.Getenv("flag"))
	}, cyberpocAuth)

	e.POST("/login", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")

		var u User
		users := getUsers()
		for _, user := range users {
			if user.Username == username {
				if md5X(password) == user.Password {
					u = user
					break
				}
			}
		}
		if u.Username == "" {
			return c.HTML(400, `username or password is wrong.`)
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

	_, err := os.Stat(sessionDir)
	if err != nil {
		if err := os.MkdirAll(sessionDir, os.ModePerm); err != nil {
			e.Logger.Fatal(err)
		}
	}

	e.Logger.Fatal(e.Start(":80"))
}

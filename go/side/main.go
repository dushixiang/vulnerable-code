package main

import (
	"bytes"
	"crypto/md5"
	realRand "crypto/rand"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const sessionDir = "sessions"

type User struct {
	Username string
	Password string
}

var user = User{
	Username: uuid.NewString(),
	Password: uuid.NewString(),
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
    <title>CyberPoC</title>
</head>
<body>
<H1>CyberPoC</H1>

<form action="/login" method="post" enctype="application/x-www-form-urlencoded">
 <label for="uname"><b>Username</b></label>
    <input type="hidden" name="csrf" value="%s" required>

    <input type="text" placeholder="Enter Username" name="username" required>

    <label for="psw"><b>Password</b></label>
    <input type="password" placeholder="Enter Password" name="password" required>
    <button type="submit">Login</button>
</form>
</body>
</html>`

func randomByte(l int) []byte {
	output := make([]byte, l)
	rand.Read(output)
	return output
}

func md5V2(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

func randomStr() string {
	return md5V2(string(randomByte(32)))
}

func main() {

	e := echo.New()
	e.Debug = true
	e.Use(middleware.Gzip())
	// 限制一秒只能发起5个请求
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(5)))

	// 模拟系统运行了一段时间
	result1, _ := realRand.Int(realRand.Reader, big.NewInt(100000))
	now := time.Now().UTC().Unix() - result1.Int64()
	rand.Seed(now)

	e.GET("/", func(c echo.Context) error {
		u, err := getUser(c)
		if err != nil {
			return c.HTML(200, fmt.Sprintf(loginHtml, randomStr()))
		}
		return c.HTML(200, `<div> Hello: `+u.Username+`</div> <br/> <a href=" ">FLAG</a >`)
	})

	e.GET("/flag", func(c echo.Context) error {
		return c.String(http.StatusOK, os.Getenv("flag"))
	}, cyberpocAuth)

	e.POST("/login", func(c echo.Context) error {
		username := c.FormValue("username")
		password := c.FormValue("password")
		csrf := c.FormValue("csrf")

		if user.Username != username || user.Password != password {
			return c.String(400, "username or password is wrong")
		}
		// 获取随机数
		result, _ := realRand.Int(realRand.Reader, big.NewInt(10*10))
		var tokens = []string{csrf}
		for i := int64(0); i < result.Int64(); i++ {
			tokens = append(tokens, randomStr())
		}
		v := strings.Join(tokens, "-")

		session := md5V2(v)
		var f = path.Join(sessionDir, session)

		file, err := os.Create(f)
		if err != nil {
			return err
		}

		encoder := gob.NewEncoder(file)
		if err := encoder.Encode(User{
			Username: username,
		}); err != nil {
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
		log.Println("用户登录成功，session=", session)
		return c.Redirect(302, "/")
	})

	// 模拟用户登录
	go func() {
		SimulatedLogin()
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-ticker.C:
				SimulatedLogin()
			}
		}
	}()

	_, err := os.Stat(sessionDir)
	if err != nil {
		if err := os.MkdirAll(sessionDir, os.ModePerm); err != nil {
			e.Logger.Fatal(err)
		}
	}

	e.Logger.Fatal(e.Start(":80"))
}

func SimulatedLogin() {
	// Create a Resty Client
	client := resty.New()

	resp, err := client.R().
		EnableTrace().
		Get("http://127.0.0.1")

	if err != nil {
		log.Fatal("get index html err", err.Error())
		return
	}

	// 使用 goquery 解析 HTML 页面
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		log.Fatal(err)
		return
	}

	// 获取 CSRF 值
	csrfValue := doc.Find("input[name='csrf']").AttrOr("value", "")
	fmt.Println("CSRF value:", csrfValue)

	formData := map[string]string{
		"username": user.Username,
		"password": user.Password,
		"csrf":     csrfValue,
	}

	resp, err = client.R().
		SetFormData(formData).
		Post(`http://127.0.0.1/login`)
	if err != nil {
		log.Fatal("login err", err.Error())
		return
	}

	result := string(resp.Body())
	if !strings.Contains(result, user.Username) {
		log.Fatal("login failed", result)
		return
	}

	fmt.Println("登录成功")
}

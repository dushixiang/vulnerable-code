package main

import (
	"bytes"
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/labstack/echo/v4"
)

// User 用户结构
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// 模拟数据库
var users = map[int64]User{}

var id atomic.Int64

// 自增ID
func incId() int64 {
	id.Add(1)
	return id.Load()
}

var guest User

func init() {
	// 初始化管理员
	adminId := incId()
	admin := User{
		ID:       adminId,
		Username: "admin",
		Password: os.Getenv("flag"),
	}
	users[adminId] = admin
	log.Println("admin username is", admin.Username)
	log.Println("admin password is", admin.Password)
	// 初始化普通用户
	userId := incId()
	guest = User{
		ID:       userId,
		Username: "guest",
		Password: "guest",
	}
	users[userId] = guest
}

//go:embed index.html
var indexHtml string

func main() {
	e := echo.New()

	// 主页
	e.GET("/", func(c echo.Context) error {
		tmpl, err := template.New("hello").Parse(indexHtml)
		if err != nil {
			log.Fatal(err)
		}
		buf := bytes.Buffer{}
		if err := tmpl.Execute(&buf, guest); err != nil {
			log.Fatal(err)
		}
		return c.HTML(http.StatusOK, buf.String())
	})

	e.GET("/api/user/:id", func(c echo.Context) error {
		userId, _ := strconv.Atoi(c.Param("id"))
		targetUser := users[int64(userId)]
		return c.JSON(http.StatusOK, targetUser)
	})
	e.Logger.Fatal(e.Start(":80"))
}

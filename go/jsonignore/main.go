package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// User 定义了用户结构体
type User struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	IsAdmin  bool   `json:"-,omitempty"` // 忽略来自前端输入的 json
}

func main() {
	e := echo.New()

	e.Debug = false
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to the Go registration system. Please POST /register to join.")
	})

	e.POST("/register", func(c echo.Context) error {
		var u = User{}

		// Echo 的 Bind 默认使用 encoding/json
		if err := c.Bind(u); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "Invalid JSON",
			})
		}

		// 检查是否注册成了管理员
		if u.IsAdmin {
			// 只有管理员才能拿 Flag
			return c.JSON(http.StatusOK, echo.Map{
				"message": "Welcome, Admin!",
				"flag":    os.Getenv("flag"),
			})
		}

		return c.JSON(http.StatusOK, echo.Map{
			"message": "Registration successful. You are a normal user.",
			"user":    u.Username,
			"isAdmin": false,
		})
	})

	e.Logger.Fatal(e.Start(":80"))
}

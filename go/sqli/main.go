package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/glebarez/sqlite"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
)

type User struct {
	ID       string
	Username string
	Password string
}

type Flag struct {
	TheFlagYouFoundMustBeMe string
}

//go:embed index.html
var indexHtml string

func main() {
	e := echo.New()
	e.Debug = true
	e.Use(middleware.Recover())

	db, err := providerDB()
	if err != nil {
		e.Logger.Fatal(err)
		return
	}

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, indexHtml)
	})
	e.POST("/login", func(c echo.Context) error {
		return login(db, c)
	})

	// Start server
	e.Logger.Fatal(e.Start(":80"))
}

func login(db *gorm.DB, c echo.Context) error {

	username := c.FormValue("username")
	password := c.FormValue("password")

	var sql = fmt.Sprintf(`SELECT username, password FROM users WHERE username=('%s') and password=('%s') LIMIT 1`, username, password)

	log.Println(sql)

	var u User
	err := db.Raw(sql).Scan(&u).Error
	if err != nil {
		return err
	}
	return c.String(200, "Landing")
}

func providerDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// 自动创建表
	if err := db.AutoMigrate(&User{}, &Flag{}); err != nil {
		return nil, err
	}
	// 插入flag
	theFlag := os.Getenv("flag")
	log.Println("insert the flag:", theFlag)
	f := &Flag{
		TheFlagYouFoundMustBeMe: theFlag,
	}
	if db.Create(f).Error != nil {
		return nil, db.Create(f).Error
	}
	return db, nil
}

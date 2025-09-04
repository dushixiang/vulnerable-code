package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewAccount(username string) *Account {
	return &Account{
		Username:      username,
		Currency:      "$",
		Balance:       10,
		LastCheckInAt: time.Now().Add(time.Minute * -1),
	}
}

type Account struct {
	Username      string
	Currency      string
	Balance       int64
	LastCheckInAt time.Time
}

func (r *Account) CheckIn() bool {
	if r.LastCheckInAt.Before(time.Now().Add(time.Minute * -1)) {
		r.Balance += 10
		r.LastCheckInAt = time.Now()
		return true
	}
	return false
}

func (r *Account) Withdraw(amount int64) int64 {
	if amount <= r.Balance {
		money := r.Balance - amount
		r.Balance = money
		return money
	}
	return 0
}

var account = NewAccount("Ethan")

//go:embed index.html
var indexHtml string

func main() {
	e := echo.New()
	e.Debug = true
	e.Use(middleware.Gzip())

	e.GET("/", func(c echo.Context) error {
		t, err := template.New("index").Parse(indexHtml)
		if err != nil {
			return err
		}
		return t.Execute(c.Response(), account)
	})

	e.GET("/withdraw", func(c echo.Context) error {
		amount, _ := strconv.ParseInt(c.QueryParam("amount"), 10, 64)
		withdraw := account.Withdraw(amount)
		return c.String(200, fmt.Sprintf("You request withdraw %s<b>%d</b>, success withdraw %s<b>%d</b>",
			account.Currency, amount, account.Currency, withdraw))
	})

	e.GET("/check-in", func(c echo.Context) error {
		checkIn := account.CheckIn()
		if checkIn {
			return c.String(200, fmt.Sprintf("You check-in <b>%v</b>, balance + $10", checkIn))
		}
		return c.String(200, fmt.Sprintf("You check-in <b>%v</b>, please after %s check-in.",
			checkIn, account.LastCheckInAt.Add(time.Minute).Format(time.RFC3339)))
	})

	e.GET("/buy", func(c echo.Context) error {
		if account.Balance > 150 {
			flag := os.Getenv("flag")
			return c.String(200, flag)
		}
		return c.String(200, "You are too poor")
	})

	e.Logger.Fatal(e.Start(":80"))
}

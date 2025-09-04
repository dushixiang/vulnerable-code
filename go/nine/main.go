package main

import (
	_ "embed"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"html/template"
	"math/rand"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

type Product struct {
	Id    int
	Word  string
	Class string
	Price uint
}

//go:embed index.html
var indexHtml string

func main() {

	var words = []rune("䷁䷖䷇䷓䷏䷢䷬䷋䷎䷳䷦䷴䷽䷷䷞䷠䷆䷃䷜䷺䷧䷿䷮䷅䷭䷑䷯䷸䷟䷱䷛䷫䷗䷚䷂䷩䷲䷔䷐䷘䷣䷕䷾䷤䷶䷝䷰䷌䷒䷨䷻䷼䷵䷥䷹䷉䷊䷙䷄䷈䷡䷍䷪䷀")
	var classes = []string{
		"bg-red-500",
		"bg-blue-500",
		"bg-green-500",
		"bg-yellow-500",
		"bg-pink-500",
		"bg-purple-500",
		"bg-orange-500",
		"bg-indigo-500",
	}

	var chooseOption = func() int64 {
		rand.Seed(time.Now().UnixNano())
		choice := rand.Intn(len(words))
		return int64(choice)
	}

	var restart = func() {
		op := chooseOption()
		atomic.StoreInt64(&option, op)
		atomic.StoreInt64(&hp, 1)
	}

	var products []Product
	for i, word := range words {
		products = append(products, Product{
			Id:    i,
			Word:  string(word),
			Class: classes[rand.Intn(8)],
			Price: 1,
		})
	}

	e := echo.New()
	e.Debug = true
	e.Use(middleware.Gzip())

	e.GET("/", func(c echo.Context) error {
		t, err := template.New("index").Parse(indexHtml)
		if err != nil {
			return err
		}
		return t.Execute(c.Response(), map[string]interface{}{
			"hp":       atomic.LoadInt64(&hp),
			"products": products,
		})
	})

	e.GET("/restart", func(c echo.Context) error {
		restart()
		return c.Redirect(302, "/")
	})

	e.GET("/choose/:choose/:price", func(c echo.Context) error {
		choose, _ := strconv.ParseInt(c.Param("choose"), 10, 64)
		price, _ := strconv.ParseInt(c.Param("price"), 10, 64)

		if atomic.LoadInt64(&hp) == 0 {
			return c.String(200, `你已经被美貌女鬼吸干了阳气，请重新开始。`)
		}
		// 扣除血量
		atomic.AddInt64(&hp, price*-1)

		if atomic.LoadInt64(&option) == choose {
			flag := os.Getenv("flag")
			return c.String(200, flag)
		}

		return c.String(200, "你选错了，代价是【被美貌女鬼吸干阳气】。")
	})

	restart()
	e.Logger.Fatal(e.Start(":80"))
}

var (
	hp     int64 = 0
	option int64 = 0
)

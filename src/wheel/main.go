package main

import (
	"math/rand"
	"fmt"
	"strings"
	"time"
	// "sync"

	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
)

// Prate ...
type Prate struct {
	Rate 	int		// 万分之N的中奖概率
	Total 	int		// 总数量限制, 0表示无限数量
	CodeA	int		// 中奖概率起始编码(包含)
	CodeB	int		// 中奖概率终止编码(包含)
	Left	int		// 剩余数
}

var prizeList []string = []string {
	"一等奖, 火星单程船票",
	"二等奖, 凉飕飕南极之旅",
	"三等奖, iphone一部",
	"",
}

var rateList []Prate = []Prate{
	Prate{Rate: 1, Total: 1, CodeA: 0, CodeB: 0, Left: 1},
	Prate{Rate: 2, Total: 2, CodeA: 1, CodeB: 2, Left: 2},
	Prate{Rate: 5, Total: 10, CodeA: 3, CodeB: 5, Left: 10},
	Prate{Rate: 100, Total: 0, CodeA: 0, CodeB: 9999, Left: 0},
}

type lotteryController struct {
	Ctx iris.Context
}

func newApp() *iris.Application {
	app := iris.New()
	mvc.New(app.Party("/")).Handle(&lotteryController{})
	return app
}

func main() {
	app := newApp()
	app.Run(iris.Addr(":8080"))
}

func (c *lotteryController) Get() string {
	c.Ctx.Header("Content-Type", "text/html")
	return fmt.Sprintf("大转盘奖品列表: <br/>%s", strings.Join(prizeList, "<br/>\n"))
}

func (c *lotteryController) GetDebug() string {
	return fmt.Sprintf("获奖概率: %v\n", rateList)
}

func (c *lotteryController) GetPrize() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := r.Intn(10000)

	var myprize string
	var prizeRate *Prate

	for i, prize := range prizeList {
		rate := &rateList[i]
		if code >= rate.CodeA && code <= rate.CodeB {
			myprize = prize
			prizeRate = rate
			break
		}
	}
	if myprize == "" {
		myprize = "很遗憾, 再来一次吧"
		return myprize
	}

	// 发奖
	if prizeRate.Total == 0 {
		return myprize
	} else if prizeRate.Left > 0 {
		prizeRate.Left --
		return myprize
	} else {
		myprize = "很遗憾, 再来一次吧"
		return myprize
	}

}
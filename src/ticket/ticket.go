package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
)

type Ticket struct {
	Ctx iris.Context
}

func NewApp() (app *iris.Application) {
	app = iris.New()
	mvc.New(app.Party("/")).Handle(&Ticket{})
	return
}
func main() {
	app := NewApp()
	// userList = []string{}
	// mu = sync.Mutex{}
	app.Run(iris.Addr(":8888"))
}

func (c *Ticket) Get() string {
	seed := time.Now().Unix()

	code := rand.New(rand.NewSource(seed)).Intn(10)
	var prize string

	switch {
	case code == 1:
		prize = "first prize"
	case code >= 2 && code <= 4:
		prize = "second prize"
	case code > 4 && code <= 6:
		prize = "third prize"
	default:
		return fmt.Sprintf("尾号为1获得一等奖<br/>"+
			"尾号为2/3/4获得二等奖<br/>"+
			"尾号为5/6获得三等奖<br>"+
			"code:%d<br/>, 很遗憾,你没有获奖", code)
	}
	return fmt.Sprintf("尾号为1获得一等奖<br/>"+
		"尾号为2/3/4获得二等奖<br/>"+
		"尾号为5/6获得三等奖<br>"+
		"code:%d<br/>, 恭喜你获得了%s", code, prize)
}

func (c *Ticket) GetPrize() string {
	seed := time.Now().Unix()
	r := rand.New(rand.NewSource(seed))
	var code [7]int

	for i := 1; i < 6; i++ {
		code[i] = r.Intn(33)
	}

	code[6] = r.Intn(13)

	return fmt.Sprintf("the codes is:%v", code)
}

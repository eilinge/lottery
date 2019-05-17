package main

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"fmt"
	"strings"
	"time"
	"math/rand"
)

type lotteryController struct {
	Ctx iris.Context
}

var userList []string

func newApp() (app *iris.Application) {
	app = iris.New()
	mvc.New(app.Party("/")).Handle(&lotteryController{})
	return 
}
func main() {
	app := newApp()
	userList = []string{}
	app.Run(iris.Addr(":8080"))
}

func (c *lotteryController) Get() string{
	count := len(userList)
	return fmt.Sprintf("online person num: %d\n", count)
}

func (c *lotteryController) PostImport() string {
	strUsers := c.Ctx.FormValue("users")
	users := strings.Split(strUsers, ",")
	count1 := len(userList)
	for _, u := range users {
		u = strings.TrimSpace(u)
		if len(u) > 0 {
			userList = append(userList, u)
		}
	}

	count2 := len(userList)
	return fmt.Sprintf("online person sum:%d, import user successfully:%d", count1, (count2-count1))
}

func(c *lotteryController) GetLucky() string {
	count := len(userList)
	if count > 1 {
		seed := time.Now().UnixNano()
		index := rand.New(rand.NewSource(seed)).Int31n(int32(count))
		user := userList[index]
		userList = append(userList[:index], userList[index+1:0]...)
		return fmt.Sprintf("the lucky man: %s, other person num: %d", user, len(userList))
	} else if count == 1{
		return fmt.Sprintf("the lucky man: %s, other person num: %d", userList[0], count)
	} else {
		return fmt.Sprintf(" no the lucky man: %s, other person num: %d", userList[0], count)
	}
}
package main

import (
	"math/rand"
	"fmt"
	"time"
	"sync"

	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
)

/*
curl http://localhost:8080/set?uid=1&menoy=100&num=20

curl http://localhost:8080/get?uid=1&id=4215640436

并发压力测试:
	wrk -t10 -c10 -d5 "http://localhost:8080/set?uid=1&menoy=100&num=20"

线程安全:
	sync.Map(>1000W)
		劣势: 多个并发时, 压力与性能不好
		优势: 大量的读, 少量写, 则压力与性能方面不明显
*/

// var packageList map[uint32][]uint
var packageList *sync.Map = new(sync.Map)

type task struct {
	id uint32
	callback chan uint
}

var chTasks chan task

const taskNum = 16
var chTaskList []chan task = make([]chan task, taskNum)

type lotteryController struct {
	Ctx iris.Context
}

func newApp() *iris.Application {
	app := iris.New()
	mvc.New(app.Party("/")).Handle(&lotteryController{})
	
	for i := 0; i < taskNum; i++ {
		chTaskList[i] = make(chan task)
		go fetchPackageListMoney(chTaskList[i])
	}
	return app
}

func main() {
	app := newApp()
	// packageList = make(map[uint32][]uint)
	chTasks = make(chan task)
	app.Run(iris.Addr(":8080"))
}

// 返回全部红包地址
// http://localhost:8080
func (c *lotteryController) Get() map[uint32][2]int {
	rs := make(map[uint32][2]int)
	// for id, list := range packageList {
	// 	var money int
	// 	for _, v := range list {
	// 		money += int(v)
	// 	}
	// 	// [2]int{钱包数, 该钱包总金额}
	// 	rs[id] = [2]int{len(list), money}
	// }
	packageList.Range(func(key, value interface{}) bool {
		// 赋值类型
		id := key.(uint32)
		list := value.([]uint)

		var money int
		for _, v := range list {
			money += int(v)
		}
		rs[id] = [2]int{len(list), money}
		return true
	})
	return rs
}

// 发红包
// http://localhost:8080/set?uid=1&menoy=100&num=10
func (c *lotteryController) GetSet() string {
	uid, errUID := c.Ctx.URLParamInt("uid")
	money, errMoney := c.Ctx.URLParamFloat64("menoy")
	num, errNum := c.Ctx.URLParamInt("num")

	if errUID != nil || errMoney != nil || errNum != nil {
		return fmt.Sprintf("参数格式异常, errUID=%d, errMoney=%d, errNum=%d\n", errUID, errMoney, errNum)
	}

	// 17.5元 = 1750分
	moneyTotal := int(money * 100)

	if uid < 1 || moneyTotal < num || num < 1 {
		return fmt.Sprintf("参数数值异常, uid=%d, money=%f, num=%d\n", uid, money, num)
	}

	// 金额分配算法
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 10元中最多5.5元
	rMax := 0.55
	if num > 1000 {
		rMax = 0.01
	} else if num >= 100 {
		rMax = 0.1
	} else if num >= 10 {
		rMax = 0.3
	}

	list := make([]uint, num)
	leftMoney := moneyTotal
	leftNum := num
	for leftNum > 0 {
		if leftNum == 1 {
			list[num-1] = uint(leftMoney)
		}else if leftMoney == leftNum {
			for i := num-leftNum; i < num; i++ {
				list[i] = 1
			}
			break
		}
		rMoney := int(float64(leftMoney-leftNum) * rMax)
		m := r.Intn(rMoney)
		if m < 1 {
			m = 1
		}
		list[num-leftNum] = uint(m)
		leftMoney -= m
		leftNum--
	}
	id := r.Uint32()
	// packageList[id] = list
	packageList.Store(id, list)
	return fmt.Sprintf("/set?id=%d&uid=%d&num=%d", id, uid, num)
}

// 抢红包
// http://localhost:8080/get?uid=1&id=1689872044
func (c *lotteryController) GetGet() string {
	uid, errUID := c.Ctx.URLParamInt("uid")
	id, errID := c.Ctx.URLParamInt("id")

	if errUID != nil || errID != nil {
		return fmt.Sprintf("")
	}
	if uid < 1 || id < 1 {
		return fmt.Sprintf("")
	}
	// list, ok := packageList[uint32(id)]

	list1, ok := packageList.Load(uint32(id))

	if !ok || list1 == nil {
		return fmt.Sprintf("红包不存在, id=%d\n", id)
	}
	// list := list1.([]uint)
	// 分配一个随机数
	// r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// i := r.Intn(len(list))

	// money := list[i]
	// // 更新红包列表中的信息
	// if len(list) > 1 {
	// 	if i == len(list)-1 {
	// 		// packageList[uint32(id)] = list[:i-1]
	// 		packageList.Store(uint32(id), list[:i-1])
	// 	} else if i == 0 {
	// 		// packageList[uint32(id)] = list[i+1:]
	// 		packageList.Store(uint32(id), list[i+1:])
	// 	} else {
	// 		// packageList[uint32(id)] = append(list[:i], list[i+1:]...)
	// 		packageList.Store(uint32(id), append(list[:i], list[i+1:]...))
	// 	}
	// } else {
	// 	// delete(packageList, uint32(id))
	// 	packageList.Delete(uint32(id))
	// }

	// 构造一个抢红包任务
	callback := make(chan uint)
	t := task{id: uint32(id), callback: callback}
	// 发送任务
	chTasks := chTaskList[id % taskNum]
	chTasks <- t
	// 接收返回结果
	// for {
	money := <-t.callback
	if money <= 0 {
		return "sorry, 没抢到红包\n"
	}
	return fmt.Sprintf("恭喜你抢到了红包, 金额为: %d\n", money)
	// }
	
} 

// 抢红包
func fetchPackageListMoney(chTasks chan task) {
	for {
		t := <- chTasks
		id := t.id
		
		l, ok := packageList.Load(uint32(id))
		if ok && l != nil {
			list := l.([]uint)

			// 分配一个随机数
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			i := r.Intn(len(list))

			money := list[i]
			// 更新红包列表中的信息
			if len(list) > 1 {
				if i == len(list)-1 {
					// packageList[uint32(id)] = list[:i-1]
					packageList.Store(uint32(id), list[:i-1])
				} else if i == 0 {
					// packageList[uint32(id)] = list[i+1:]
					packageList.Store(uint32(id), list[1:])
				} else {
					// packageList[uint32(id)] = append(list[:i], list[i+1:]...)
					packageList.Store(uint32(id), append(list[:i], list[i+1:]...))
				}
			} else {
				// delete(packageList, uint32(id))
				packageList.Delete(uint32(id))
			}
			t.callback <- money
		}
		t.callback <- 0
	}
}
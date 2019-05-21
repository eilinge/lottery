package main

/*
压力测试:
	wrk -t10 -c10 -d5 http://localhost:8080/lucky
	减少奖品的数量, 能够提升性能--抽奖(lucky)占用大量开销
	性能:
		Requests/sec:  13904.00
		Transfer/sec:      2.03MB
保证多请求时, 线程不安全, sync.Mutex 加锁:
	性能:
		Requests/sec:   5050.09
		Transfer/sec:      0.96MB
*/
import (
	"log"
	"os"
	"fmt"
	"time"
	"math/rand"
	"sync"

	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
)

const (
	giftTypeCoin = iota	// 虚拟币
	giftTypeCoupon		// 不同券
	giftTypeCouponFix	// 相同的券
	giftTypeRealSmall	// 实物小奖
	giftTypeRealLarge	// 实物大奖
)

type gift struct {
	id 			int
	name 		string 	// 奖品名称
	pic 		string 	// 奖品的图片
	link 		string 	// 奖品的链接
	gtype 		int 	// 奖品类型
	data 		string 	// 奖品的数据
	datalist[] 	string 	// 奖品数据集合(不同的优惠券的编码)
	total 		int 	// 总数, 0 不限量
	left 		int 	// 剩余数量
	inuse 		bool 	// 是否使用中
	rate 		int 	// 中奖概率
	rateMin 	int 	// 大于等于最小中奖编码
	rateMax 	int 	// 小于中奖编码
}

const rateMax = 10000

var mu sync.Mutex

var logger *log.Logger

var giftList []*gift

type lotteryController struct {
	Ctx *iris.Context
}

func initLog() {
	f, err := os.Create("./lotttery_demo.log")
	if err != nil {
		fmt.Println("os.Create file err: ", err)
		return
	}
	logger = log.New(f, "", log.Ldate|log.Lmicroseconds)
}

func initGift() {
	giftList = make([]*gift, 5)
	giftList[0] = &gift{
		id:	1,
		name:	"iphoneX",
		pic: "",
		link: "",
		gtype:	giftTypeRealLarge,
		data: "",
		datalist: nil,
		total: 10000,
		left: 10000,
		inuse: true,
		rate: 10000,
		rateMin: 0,
		rateMax: 0,
	}

	giftList[1] = &gift{
		id:	2,
		name:	"充电器",
		pic: "",
		link: "",
		gtype:	giftTypeRealSmall,
		data: "",
		datalist: nil,
		total: 5,
		left: 5,
		inuse: false,
		rate: 100,
		rateMin: 0,
		rateMax: 0,
	}

	giftList[2] = &gift{
		id:	3,
		name:	"优惠券满200减50元",
		pic: "",
		link: "",
		gtype:	giftTypeCouponFix,
		data: "small-coupon-2018",
		datalist: nil,
		total: 50,
		left: 50,
		inuse: false,
		rate: 100,
		rateMin: 0,
		rateMax: 0,
	}

	giftList[3] = &gift{
		id:	4,
		name:	"直降优惠券",
		pic: "",
		link: "",
		gtype:	giftTypeCoupon,
		data: "",
		datalist: []string{"c01", "c02", "c03", "c04", "c05"},
		total: 50,
		left: 50,
		inuse: false,
		rate: 200,
		rateMin: 0,
		rateMax: 0,
	}

	giftList[4] = &gift{
		id:	5,
		name:	"金币",
		pic: "",
		link: "",
		gtype:	giftTypeCoin,
		data: "10金币",
		datalist: nil,
		total: 40,
		left: 40,
		inuse: false,
		rate: 2000,
		rateMin: 0,
		rateMax: 0,
	}
	// 数据整理, 中奖区间数据
	rateStart := 0
	for _, data := range giftList {
		if !data.inuse {
			continue
		}
		data.rateMin = rateStart
		data.rateMax = rateStart + data.rate
		if data.rateMax >= rateMax {
			// 保证中奖率百分百
			data.rateMax = rateMax
			rateStart = 0
		} else {
			// 对中奖率进行累加, 得出百分比
			rateStart += data.rate
		}
	}
}

func newApp() *iris.Application{
	app := iris.New()
	mvc.New(app.Party("/")).Handle(&lotteryController{})
	initLog()
	initGift()
	return app
}

func main() {
	app := newApp()
	app.Run(iris.Addr(":8080"))
}

func (c *lotteryController) Get() string {
	count := 0
	total := 0
	for _, data := range giftList {
		if data.inuse && (data.total == 0 || (data.total > 0 && data.left > 0)) {
			count ++
			total += data.left
		}
	}
	return fmt.Sprintf("当前有效奖品种类数量: %d, 限量奖品总数量: %d", count, total)
}

func sendCoin(data *gift) (bool, string) {
	if data.total == 0 {
		return true, data.data
	} else if data.left > 0 {
		data.left = data.left - 1
		return true, data.data
	} else {
		return false, "奖品已发完"
	}
}

func sendCoupon(data *gift) (bool, string) {
	if data.left > 0 && data.left < len(data.datalist){
		left := data.left - 1
		data.left = left
		return true, data.datalist[left]
	}
	return false, "奖品已发完"

}

func (c *lotteryController) GetLucky() map[string]interface{} {
	mu.Lock()
	defer mu.Unlock()

	code := luckCode()
	ok := false
	result := make(map[string]interface{})
	result["success"] = ok
	for _, data := range giftList {
		// 不能使用, 无库存
		if !data.inuse || (data.total > 0 && data.left <= 0 ) {
			continue
		}
		if data.rateMin <= int(code) && data.rateMax > int(code) {
			// 中奖了, 抽奖编码在奖品编码范围内
			// 开始发奖
			sendData := ""
			switch data.gtype {
			case giftTypeCoin:
				ok, sendData = sendCoin(data)
			case giftTypeCoupon:
				ok, sendData = sendCoupon(data)
			case giftTypeCouponFix:
				ok, sendData = sendCoin(data)
			case giftTypeRealSmall:
				ok, sendData = sendCoin(data)
			case giftTypeRealLarge:
				ok, sendData = sendCoin(data)
			}
			if ok {
				saveLuckyData(code, data.id, data.name, data.link, sendData, data.left)
				result["success"] = ok
				result["id"] = data.id
				result["name"] = data.name
				result["link"] = data.link
				result["data"] = sendData
				break
			}

		}
	}
	return result
}

func luckCode() int32 {
	seed := time.Now().Unix()
	return rand.New(rand.NewSource(seed)).Int31n(int32(rateMax))
}

// 记录用户获奖信息
func saveLuckyData(code int32, id int, name, link, sendData string, left int) {
	logger.Printf("lucky, code: %d, gift: %d, name: %s, link: %s, sendData: %s, left: %d",
		code, id, name, link, sendData, left)
}
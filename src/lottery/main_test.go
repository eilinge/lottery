package main

import (
	"fmt"
	"sync"
	"testing"

	"github.com/kataras/iris/httptest"
	_ "github.com/iris-contrib/httpexpect"
)

func TestMVC(t *testing.T) {
	e := httptest.New(t, NewApp())

	var wg sync.WaitGroup // 协程同步, 保证协程全部运行完成

	e.GET("/").Expect().Status(httptest.StatusOK).
		Body().Equal("\nonline person num: 0\n")

	for i := 0; i < 2000000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			e.POST("/import").WithFormField("users", fmt.Sprintf("test_u%d", i)).Expect().
				Status(httptest.StatusOK)
		}(i)
	}

	wg.Wait() // 保证协程全部运行完成

	e.GET("/").Expect().Status(httptest.StatusOK).
		Body().Equal("\nonline person num: 2000000\n")
	e.GET("/lucky").Expect().Status(httptest.StatusOK)

	e.GET("/").Expect().Status(httptest.StatusOK).
		Body().Equal("\nonline person num: 1999999\n")
}

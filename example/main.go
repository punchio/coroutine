package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/punchio/coroutine"
)

func printTime(params ...interface{}) {
	fmt.Println(time.Now().String(), params)
}

func DoSomething(cmd string, d time.Duration) func() (interface{}, error) {
	return func() (interface{}, error) {
		<-time.After(d)
		return cmd + " " + d.String(), nil
	}
}

func taskFunc(co *coroutine.Task) interface{} {
	// printTime("task start")
	co.Wait(time.Second)
	// printTime("yield wait 1 second")

	_, _ = co.Yield(DoSomething("wait 1 second", time.Second))
	// printTime("yield return1:", data)

	//always fail
	_, _ = co.YieldWithTimeOut(DoSomething("wait timeout 3 second", time.Second*3), time.Second*2)
	// if err != nil {
	// 	// printTime("yield return6 fail,err:", err, data2)
	// } else {
	// 	// printTime("yield return6:", data2.(string))
	// }

	strs := []string{"a", "b", "c"}
	for _, v := range strs {
		_, _ = co.Yield(DoSomething(v, time.Second))
		// printTime("yield return1:", data)
	}
	return nil
}

func test(task *coroutine.Task) interface{} {
	_, _ = task.Yield(func() (interface{}, error) {
		sum := 0
		for i := 0; i < 1000000; i++ {
			sum += i
		}
		return sum, nil
	})
	return nil
}

func main() {
	go func() {
		_ = http.ListenAndServe("127.0.0.1:8899", nil)
	}()

	printTime("start")
	cg := coroutine.New()
	for i := 0; i < 500000; i++ {
		cg.Add(test)
	}

	printTime("init finish")

	for {
		//执行主线程函数
		// printTime("main run")

		//执行协程函数
		cg.Run()
		// printTime("main run finish")
		if cg.Len() == 0 {
			// printTime("test finish")
			break
		}
		<-time.After(time.Millisecond * 20)
	}
	printTime("exit")
	<-time.After(10)
}

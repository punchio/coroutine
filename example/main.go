package main

import (
	"fmt"
	"github.com/punchio/coroutine"
	"time"
)

func DoSomething(cmd string, d time.Duration) func() interface{} {
	return func() interface{} {
		<-time.After(d)
		return cmd + " " + d.String()
	}
}

func taskFunc(co *coroutine.Task) interface{} {
	fmt.Println("task start")
	co.Wait(time.Second)
	fmt.Println("yield wait 1 second")

	data := co.Yield(DoSomething("wait 1 second", time.Second)).(string)
	fmt.Println("yield return1:", data)

	data = co.Yield(DoSomething("wait 2 second", time.Second*2)).(string)
	fmt.Println("yield return2:", data)

	data = co.Yield(DoSomething("wait 3 second", time.Second*3)).(string)
	fmt.Println("yield return3:", data)

	//always success
	data2, err := co.YieldWithTimeOut(DoSomething("wait timeout 3 second", time.Second*3), time.Second*4)
	if err != nil {
		fmt.Println("yield return4 fail,err:", err, data2)
	} else {
		fmt.Println("yield return4:", data2.(string))
	}

	//maybe success,maybe fail
	data2, err = co.YieldWithTimeOut(DoSomething("wait timeout 3 second", time.Second*3), time.Second*3)
	if err != nil {
		fmt.Println("yield return5 fail,err:", err, data2)
	} else {
		fmt.Println("yield return5:", data2.(string))
	}

	//always fail
	data2, err = co.YieldWithTimeOut(DoSomething("wait timeout 3 second", time.Second*3), time.Second*2)
	if err != nil {
		fmt.Println("yield return6 fail,err:", err, data2)
	} else {
		fmt.Println("yield return6:", data2.(string))
	}
	return nil
}

func main() {
	cg := coroutine.NewCoroutine()
	cg.Add(taskFunc).OnComplete(func(data interface{}, err error) {
		fmt.Println("OnComplete data:", data, "err:", err)
	})

	for {
		//执行主线程函数
		// fmt.Println("main run")
		//执行协程函数

		cg.Run()
		// fmt.Println("main run finish")
		if cg.Len() == 0 {
			fmt.Println("test finish")
			break
		}
	}
}

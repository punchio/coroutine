package main

import (
	"fmt"
	"time"

	"github.com/punchio/coroutine"
)

func printTime(params ...interface{}) {
	fmt.Println(time.Now().String(), params)
}

func DoSomething(cmd string, d time.Duration) func() interface{} {
	return func() interface{} {
		<-time.After(d)
		return cmd + " " + d.String()
	}
}

func taskFunc(task *coroutine.Task) interface{} {
	printTime("task start")
	task.Wait(time.Second)
	printTime("yield wait: 1 second")

	data := task.Yield(DoSomething("wait 1 second", time.Second)).(string)
	printTime("yield return1:", data)

	//always fail
	data2, err := task.YieldWithTimeOut(DoSomething("wait timeout 3 second", time.Second*3), time.Second*2)
	if err != nil {
		printTime("yield return6 fail,err:", err, data2)
	} else {
		printTime("yield return6:", data2.(string))
	}

	strs := []string{"a", "b", "c"}
	for _, v := range strs {
		data = task.Yield(DoSomething(v, time.Second)).(string)
		printTime("yield return1:", data)
	}
	return nil
}

//模拟消息处理
var registers map[int]func(*coroutine.Task, interface{}) error

func reg(msgType int, f func(*coroutine.Task, interface{}) error) {
	registers[msgType] = f
}

func onMsgAsync(co *coroutine.Coroutine, msgType int, data interface{}) {
	f := registers[msgType]
	co.Add(func(task *coroutine.Task) interface{} {
		return f(task, data)
	}).OnComplete(func(data interface{}, err error) {
		printTime("result:", data, "err:", err)
	})
}

func onSysEnter(task *coroutine.Task, data interface{}) error {
	d := data.(time.Duration)
	printTime("onSysEnter:", d.String())
	_, err := task.YieldWithTimeOut(DoSomething("", d), time.Second)
	if err != nil {
		return err
	}
	r := task.Yield(reqRpc(1, 2))
	switch data := r.(type) {
	case error:
		printTime("err:", err)
	case interface{}:
		printTime("rpc:", data)
	}
	return nil
}

//rpc处理
func sendRpc(target int, req int, cb func(data interface{}, err error)) {
	go func() {
		<-time.After(time.Second * 2)
		cb("rpc", nil)
	}()
}

func reqRpc(target int, req int) func() interface{} {
	recv := make(chan interface{})
	cb := func(data interface{}, err error) {
		if err != nil {
			recv <- err
		} else {
			recv <- data
		}
	}

	sendRpc(target, req, cb)
	return func() interface{} {
		return <-recv
	}
}

func main() {
	registers = make(map[int]func(*coroutine.Task, interface{}) error)
	reg(1, onSysEnter)

	cg := coroutine.NewCoroutine()
	cg.Add(taskFunc).OnComplete(func(data interface{}, err error) {
		printTime("OnComplete data:", data, "err:", err)
	})

	onMsgAsync(cg, 1, time.Millisecond)
	onMsgAsync(cg, 1, time.Minute)

	for {
		//执行主线程函数
		// printTime("main run")

		//执行协程函数
		cg.Run()
		// printTime("main run finish")
		if cg.Len() == 0 {
			printTime("test finish")
			break
		}
	}
}

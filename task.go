package coroutine

import (
	"fmt"
	"time"
)

type TaskEntry func(*Task) interface{}
type TaskYieldFunction func() (interface{}, error)
type TaskCompletion func(interface{}, error)

type Task struct {
	//主控协程
	co *Coroutine
	//主线程与协程同步信号
	signal chan struct{}

	//完成结果
	onComplete TaskCompletion
	done       bool
	result     interface{}
	err        error
}

//创建
func newTask(co *Coroutine, completion TaskCompletion) *Task {
	t := &Task{}
	t.co = co
	t.signal = make(chan struct{})
	t.onComplete = completion
	return t
}

func (self *Task) finish() {
	self.done = true
	if self.onComplete != nil {
		self.onComplete(self.result, self.err)
	}
	self.co.enqueueTask(self)
	self.co.resume()
}

//只能在协程调用
func (self *Task) yield(f TaskYieldFunction) (interface{}, error) {
	//f为空，直接返回
	if f == nil {
		return nil, nil
	}
	// fmt.Println("yield enter")

	//交出主线程控制权
	self.co.resume()
	//执行阻塞操作
	data, err := f()
	// fmt.Println("yield finish")

	//执行完成进入就绪队列，阻塞等待主线程唤醒
	self.co.enqueueTask(self)
	<-self.signal
	// fmt.Println("yield exit")
	return data, err
}

//只能在主线程调用
func (self *Task) resume() {
	// fmt.Println("resume enter")
	self.signal <- struct{}{}
	self.co.yield()
	// fmt.Println("resume exit")
}

func (self *Task) handlePanic() {
	e := recover()
	if e != nil {
		switch err := e.(type) {
		case error:
			self.err = fmt.Errorf("task panic recovery with error: %s", err.Error())
		default:
			self.err = fmt.Errorf("task panic recovery with unknown error: %s", fmt.Sprint(err))
		}
	}
}

func (self *Task) Yield(f TaskYieldFunction) (interface{}, error) {
	return self.yield(f)
}

func (self *Task) Wait(d time.Duration) {
	_, _ = self.yield(func() (interface{}, error) {
		<-time.After(d)
		return nil, nil
	})
}

func (self *Task) YieldWithTimeOut(f TaskYieldFunction, wait time.Duration) (interface{}, error) {
	if f == nil {
		return nil, nil
	}
	var data interface{}
	var err error
	c := make(chan struct{})
	go func() {
		data, err = f()
		c <- struct{}{}
	}()

	return self.yield(func() (interface{}, error) {
		select {
		case data = <-c:
			// fmt.Println("receive data")
		case <-time.After(wait):
			// fmt.Println("task timeout")
			err = fmt.Errorf("task timeout")
		}

		return data, err
	})

}

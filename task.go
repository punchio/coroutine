package coroutine

import (
	"fmt"
	"sync"
	"time"
)

type TaskYieldFunction func() (interface{}, error)
type TaskCompleteFunction func(interface{}, error)

type Task struct {
	//控制与主线程同步信号
	wgCo   sync.WaitGroup
	wgTask sync.WaitGroup
	//执行完成进入就绪队列
	readyChan chan<- *Task

	//处理结束回调
	onComplete TaskCompleteFunction

	//完成结果
	done   bool
	result interface{}
	err    error
}

func newTask(exe func(*Task) interface{}, ch chan<- *Task) *Task {
	t := &Task{}
	t.wgTask.Add(1)
	t.readyChan = ch
	go func() {
		defer func() {
			t.handlePanic()
			// fmt.Println("self unlock")
			//协程结束，释放控制
			t.readyChan <- t
			// fmt.Println("task finish!")
			t.done = true
			t.wgCo.Done()
			// fmt.Println("finish!")
		}()

		// fmt.Println("self lock")
		//暂停协程，等待主线程调用
		t.wgTask.Wait()
		t.result = exe(t)
	}()
	t.resume()
	// fmt.Println("new exit")
	return t
}

//只能在coroutine线程调用
func (self *Task) yield(f TaskYieldFunction) (interface{}, error) {
	// fmt.Println("yield enter")
	self.wgTask.Add(1)
	self.wgCo.Done()
	var data interface{}
	var err error
	if f != nil {
		data, err = f()
	}
	// fmt.Println("yield finish")
	self.readyChan <- self
	self.wgTask.Wait()
	// fmt.Println("yield exit")
	return data, err
}

//只能在coroutine管理线程调用
func (self *Task) resume() {
	// fmt.Println("resume enter")
	self.wgCo.Add(1)
	self.wgTask.Done()
	self.wgCo.Wait()
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

func (self *Task) OnComplete(onComplete TaskCompleteFunction) *Task {
	self.onComplete = onComplete
	return self
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

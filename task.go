package coroutine

import (
	"fmt"
	"sync"
	"time"
)

type Task struct {
	//控制与主线程同步信号
	wgCo   sync.WaitGroup
	wgTask sync.WaitGroup

	//处理结束回调
	onComplete func(data interface{}, err error)

	//完成结果
	done   bool
	result interface{}
	err    error
}

func newTask(exe func(*Task) interface{}) *Task {
	t := &Task{}
	t.wgTask.Add(1)
	go func() {
		defer func() {
			t.handlePanic()
			// fmt.Println("self unlock")
			t.done = true
			//协程结束，释放控制
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
func (self *Task) yield(f func() interface{}) interface{} {
	// fmt.Println("yield enter")
	self.wgTask.Add(1)
	self.wgCo.Done()
	data := f()
	self.wgTask.Wait()
	// fmt.Println("yield exit")
	return data
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
		case nil:
			self.err = fmt.Errorf("panic recovery with nil error")
		case error:
			self.err = fmt.Errorf("panic recovery with error: %s", err.Error())
		default:
			self.err = fmt.Errorf("panic recovery with unknown error: %s", fmt.Sprint(err))
		}
	}
}

func (self *Task) OnComplete(onComplete func(interface{}, error)) *Task {
	self.onComplete = onComplete
	return self
}

func (self *Task) Yield(f func() interface{}) interface{} {
	return self.yield(f)
}

func (self *Task) Wait(d time.Duration) {
	self.yield(func() interface{} {
		<-time.After(d)
	})
}

func (self *Task) YieldWithTimeOut(f func() interface{}, wait time.Duration) (interface{}, error) {
	timer := time.After(wait)
	c := make(chan interface{})
	go func() {
		c <- f()
	}()

	for {
		// fmt.Println("Yield")
		select {
		case data := <-c:
			// fmt.Println("receive data")
			return data, nil
		case <-timer:
			// fmt.Println("task timeout")
			return nil, fmt.Errorf("task timeout")
		default:
			// fmt.Println("no data, sleep")
			self.yield()
		}
	}
}

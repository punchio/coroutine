package coroutine

import (
	"fmt"
	"time"
)

type Coroutine struct {
	//控制与主线程同步信号
	sigCo   chan struct{}
	sigMain chan struct{}

	//处理结束回调
	onComplete func(data interface{}, err error)

	//完成结果
	done   bool
	result interface{}
	err    error
}

func newCoroutine(exe func(*Coroutine) interface{}) *Coroutine {
	t := &Coroutine{}
	t.sigMain = make(chan struct{})
	t.sigCo = make(chan struct{})
	go func() {
		defer func() {
			t.handlePanic()
			// fmt.Println("self unlock")
			t.done = true
			//协程结束，释放控制
			t.sigMain <- struct{}{}
			// fmt.Println("finish!")
		}()

		// fmt.Println("self lock")
		//暂停协程，等待主线程调用
		<-t.sigCo
		t.result = exe(t)
	}()
	t.resume()
	// fmt.Println("new exit")
	return t
}

func async(f func() interface{}) <-chan interface{} {
	c := make(chan interface{})
	go func() {
		c <- f()
	}()
	return c
}

//只能在coroutine线程调用
func (self *Coroutine) yield() {
	// fmt.Println("yield enter")
	self.sigMain <- struct{}{}
	<-self.sigCo
	// fmt.Println("yield exit")
}

//只能在coroutine管理线程调用
func (self *Coroutine) resume() {
	// fmt.Println("resume enter")
	self.sigCo <- struct{}{}
	<-self.sigMain
	// fmt.Println("resume exit")
}

func (self *Coroutine) handlePanic() {
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

func (self *Coroutine) OnComplete(onComplete func(data interface{}, err error)) *Coroutine {
	self.onComplete = onComplete
	return self
}

func (self *Coroutine) Yield(f func() interface{}) interface{} {
	c := async(f)
	for {
		// fmt.Println("Yield")
		select {
		case data := <-c:
			// fmt.Println("receive data")
			return data
		default:
			// fmt.Println("no data, sleep")
			self.yield()
		}
	}
}

func (self *Coroutine) YieldWithTimeOut(f func() interface{}, wait time.Duration) (interface{}, error) {
	c := async(f)
	timer := time.After(wait)
	for {
		// fmt.Println("Yield")
		select {
		case data := <-c:
			// fmt.Println("receive data")
			return data, nil
		case <-timer:
			fmt.Println("task timeout")
			return nil, fmt.Errorf("task timeout")
		default:
			// fmt.Println("no data, sleep")
			self.yield()
		}
	}
}

func (self *Coroutine) Wait(d time.Duration) {
	timer := time.After(d)
	for {
		// fmt.Println("Wait")
		select {
		case <-timer:
			// fmt.Println("timer expired")
			return
		default:
			// fmt.Println("no data, sleep")
			self.yield()
		}
	}
}

package coroutine

import (
	"fmt"
	"time"
)

type Coroutine struct {
	sigSelf    chan struct{}
	sigMain    chan struct{}
	onComplete func(data interface{}, err error)
	done       bool
	result     interface{}
	err        error
}

func newCoroutine(exe func(*Coroutine) interface{}) *Coroutine {
	t := &Coroutine{}
	t.sigMain = make(chan struct{})
	t.sigSelf = make(chan struct{})
	go func() {
		defer func() {
			t.handlePanic()
			// fmt.Println("self unlock")
			t.sigMain <- struct{}{}
			if t.onComplete != nil {
				t.onComplete(t.result, t.err)
			}
			t.done = true
			// fmt.Println("finish!")
		}()

		// fmt.Println("self lock")
		<-t.sigSelf
		t.result = exe(t)
	}()
	t.resume()
	// fmt.Println("new exit")
	return t
}

//只能在coroutine线程调用
func (self *Coroutine) yield() {
	// fmt.Println("yield enter")
	self.sigMain <- struct{}{}
	<-self.sigSelf
	// fmt.Println("yield exit")
}

//只能在coroutine管理线程调用
func (self *Coroutine) resume() {
	// fmt.Println("resume enter")
	self.sigSelf <- struct{}{}
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

func (self *Coroutine) Yield(c <-chan interface{}) interface{} {
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

func (self *Coroutine) YieldWithTimeOut(c <-chan interface{}, wait time.Duration) (interface{}, error) {
	timer := time.After(wait)
	for {
		// fmt.Println("Yield")
		select {
		case data := <-c:
			// fmt.Println("receive data")
			return data, nil
		default:
			if len(timer) > 0 {
				fmt.Println("task timeout")
				return nil, fmt.Errorf("task timeout")
			}

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

package coroutine

import (
	"fmt"
	"time"
)

type Coroutine struct {
	sigMain chan<- interface{}
	sigSelf chan interface{}
	done    bool
	err     error
}

func NewCoroutine(f func(*Coroutine) error, sigMain chan<- interface{}) *Coroutine {
	t := &Coroutine{}
	t.sigMain = sigMain
	t.sigSelf = make(chan interface{})
	go func() {
		<-t.sigSelf
		t.err = f(t)
		t.done = true
		t.yield()
		// fmt.Println("TaskFinish")
	}()
	return t
}

func (self *Coroutine) Done() (bool, error) {
	return self.done, self.err
}

func (self *Coroutine) yield() {
	self.sigMain <- nil
	<-self.sigSelf
}

func (self *Coroutine) Resume() {
	// fmt.Println("Task Awake")
	self.sigSelf <- nil
	// fmt.Println("Task Awaked")
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
		// fmt.Println("Yield")
		select {
		case <-timer:
			// fmt.Println("receive data")
			return
		default:
			// fmt.Println("no data, sleep")
			self.yield()
		}
	}
}

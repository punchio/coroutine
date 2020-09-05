package coroutine

import "fmt"

type CoroutineGroup struct {
	sigMain    chan interface{}
	coList     []*Coroutine
	appendList []*Coroutine
}

func NewCoroutineGroup(sig chan interface{}) *CoroutineGroup {
	cg := &CoroutineGroup{}
	cg.sigMain = sig
	return cg
}

func (self *CoroutineGroup) Add(f func(*Coroutine) error) {
	co := NewCoroutine(f, self.sigMain)
	self.appendList = append(self.appendList, co)
}

func (self *CoroutineGroup) Run() {
	for _, v := range self.coList {
		v.Resume()
		<-self.sigMain
	}

	for i := 0; i < len(self.coList); {
		if done, err := self.coList[i].Done(); done {
			self.coList = append(self.coList[:i], self.coList[i+1:]...)
			fmt.Println("Done task, err:", i, err, len(self.coList))
		} else {
			i++
		}
	}
	self.coList = append(self.coList, self.appendList...)
	self.appendList = nil
}

func (self *CoroutineGroup) Len() int {
	return len(self.appendList) + len(self.coList)
}

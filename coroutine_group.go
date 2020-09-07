package coroutine

type CoroutineGroup struct {
	coList     []*Coroutine
	appendList []*Coroutine
}

func NewCoroutineGroup() *CoroutineGroup {
	cg := &CoroutineGroup{}
	return cg
}

func (self *CoroutineGroup) Add(f func(*Coroutine) interface{}) *Coroutine {
	co := newCoroutine(f)
	self.appendList = append(self.appendList, co)
	return co
}

func (self *CoroutineGroup) Run() {
	for _, v := range self.coList {
		v.resume()
		// fmt.Println("run:", i)
	}

	for i := 0; i < len(self.coList); {
		co := self.coList[i]
		if co.done {
			if co.onComplete != nil {
				co.onComplete(co.result, co.err)
			}
			self.coList = append(self.coList[:i], self.coList[i+1:]...)
			// fmt.Println("Done task, err:", i)
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

package coroutine

type Coroutine struct {
	taskList       []*Task
	appendTaskList []*Task
}

//创建一个协程运行管理
func NewCoroutine() *Coroutine {
	cg := &Coroutine{}
	return cg
}

//添加协程
func (self *Coroutine) Add(f func(*Task) interface{}) *Task {
	co := newTask(f)
	self.appendTaskList = append(self.appendTaskList, co)
	return co
}

func (self *Coroutine) Run() {
	for _, v := range self.taskList {
		v.resume()
		// fmt.Println("run:", i)
	}

	for i := 0; i < len(self.taskList); {
		co := self.taskList[i]
		if co.done {
			if co.onComplete != nil {
				co.onComplete(co.result, co.err)
			}
			self.taskList = append(self.taskList[:i], self.taskList[i+1:]...)
			// fmt.Println("Done task, err:", i)
		} else {
			i++
		}
	}
	self.taskList = append(self.taskList, self.appendTaskList...)
	self.appendTaskList = nil
}

func (self *Coroutine) Len() int {
	return len(self.appendTaskList) + len(self.taskList)
}

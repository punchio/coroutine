package coroutine

type Coroutine struct {
	readyChan    chan *Task
	runningCount int
}

//创建一个协程运行管理
func New(chanCount int) *Coroutine {
	co := &Coroutine{}
	co.readyChan = make(chan *Task, chanCount)
	return co
}

//添加协程
func (self *Coroutine) Add(f func(*Task) interface{}) *Task {
	co := newTask(f, self.readyChan)
	self.runningCount++
	return co
}

func (self *Coroutine) Run() {
	for {
		select {
		case task := <-self.readyChan:
			if task.done {
				// fmt.Println("coroutine done")
				self.runningCount--
				if task.onComplete != nil {
					task.onComplete(task.result, task.err)
				}
				continue
			}
			// fmt.Println("coroutine resume")
			task.resume()
		default:
			return
		}
	}
}

func (self *Coroutine) Len() int {
	return self.runningCount
}

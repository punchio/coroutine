package coroutine

//单线程处理
type Coroutine struct {
	//就绪任务队列
	readyList chan *Task
	//主线程与协程同步信号
	signal chan struct{}
	//运行任务数量
	runningCount int
}

//创建一个协程管理任务
func New(chanCount int) *Coroutine {
	co := &Coroutine{}
	co.readyList = make(chan *Task, chanCount)
	co.signal = make(chan struct{})
	return co
}

//添加协程
func (self *Coroutine) Add(entry TaskEntry, completion TaskCompletion) *Task {
	self.runningCount++
	t := newTask(self, completion)
	go func() {
		defer func() {
			t.handlePanic()
			//协程结束，释放控制
			t.finish()
			// fmt.Println("task finish!")
		}()

		// fmt.Println("self lock")
		//暂停协程，等待主线程调用
		<-t.signal
		t.result = entry(t)
	}()
	t.resume()
	return t
}

func (self *Coroutine) yield() {
	<-self.signal
}

func (self *Coroutine) resume() {
	self.signal <- struct{}{}
}

func (self *Coroutine) enqueueTask(t *Task) {
	self.readyList <- t
}

func (self *Coroutine) Run() {
	for {
		select {
		case task := <-self.readyList:
			if task.done {
				// fmt.Println("coroutine done")
				self.runningCount--
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

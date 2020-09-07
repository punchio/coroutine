# golang-coroutine
模仿C#的协程，模拟async to sync
## go受控协程
让异步调用书写流程与同步一样，解决异步调用代码打断的问题。
协程线程与主线程都是受控制，通过chan来控制同步，主线程与协程不会同时执行，保证了全局变量安全性。
通过resume来恢复协程线程，让出主线程控制权
通过Yield来打断协程线程的执行，让出控制权给主线程
Yield可以返回等待数据

## 协程使用
### 1.创建
```golang
func NewCoroutine() *Coroutine
```
通过此函数创建一个协程组，主线程来控制协程的运行

### 2.运行
```golang
func (self *Coroutine) Run()
```
主线程调用Run让协程的运行起来

### 3.添加任务
```golang
func (self *Coroutine) Add(f func(*Task) interface{}) *Task
```
添加任务到协程组里面

## 任务使用
### 1.任务创建
```golang
func newTask(exe func(*Task) interface{}) *Task
```
创建一个Task，需要一个接受*Task为参数，返回一个interface{}的函数

### 2.任务暂停
```golang
func (self *Task) Yield(f func() interface{}) interface{}
func (self *Task) YieldWithTimeOut(f func() interface{}, wait time.Duration) (interface{}, error)
func (self *Task) Wait(d time.Duration)
```
在函数执行时，有以下三种可以暂停任务的方式

# 代码示例
```golang
func taskFunc(co *coroutine.Task) interface{} {
	fmt.Println("task start")
	co.Wait(time.Second)
	fmt.Println("yield wait 1 second")

	data := co.Yield(DoSomething("wait 1 second", time.Second)).(string)
	fmt.Println("yield return1:", data)

	//always fail
	data2, err := co.YieldWithTimeOut(DoSomething("wait timeout 3 second", time.Second*3), time.Second*2)
	if err != nil {
		fmt.Println("yield return6 fail,err:", err, data2)
	} else {
		fmt.Println("yield return6:", data2.(string))
	}

	strs := []string{"a", "b", "c"}
	for _, v := range strs {
		data = co.Yield(DoSomething(v, time.Second)).(string)
		fmt.Println("yield return1:", data)
	}
	return nil
}

func main() {
	cg := coroutine.NewCoroutine()
	cg.Add(taskFunc).OnComplete(func(data interface{}, err error) {
		fmt.Println("OnComplete data:", data, "err:", err)
	})

	for {
		//执行主线程函数
		// fmt.Println("main run")

		//执行协程函数
		cg.Run()
		// fmt.Println("main run finish")
		if cg.Len() == 0 {
			fmt.Println("test finish")
			break
		}
	}
}
```

# golang-coroutine
## go语言可控制协程
让异步调用书写流程与同步一样，解决异步调用代码打断的问题。
通过chan来控制与主线程同步，同时通过chan返回等待数据，所以我们的Yield都需要传入<-chan interface{}去检查数据是否返回，如果只是等待时间可以不用传入chan。
sigMain控制了主线程与协程不会同时执行，保证了全局变量安全性。
## 使用
### 1.创建
func NewCoroutine(f func(*Coroutine) error, sigMain chan<- interface{}) *Coroutine
### 2.恢复
func (self *Coroutine) Resume()
### 3.挂起
func (self *Coroutine) Yield(c <-chan interface{}) interface{}
### 4.结束
func (self *Coroutine) Done() (bool, error)

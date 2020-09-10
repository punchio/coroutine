package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/punchio/coroutine"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func printTime(params ...interface{}) {
	fmt.Println(time.Now().String(), params)
}

func DoSomething(cmd string, d time.Duration) func() (interface{}, error) {
	return func() (interface{}, error) {
		<-time.After(d)
		return cmd + " " + d.String(), nil
	}
}

func taskFunc(co *coroutine.Task) interface{} {
	// printTime("task start")
	co.Wait(time.Second)
	// printTime("yield wait 1 second")

	_, _ = co.Yield(DoSomething("wait 1 second", time.Second))
	// printTime("yield return1:", data)

	//always fail
	_, _ = co.YieldWithTimeOut(DoSomething("wait timeout 3 second", time.Second*3), time.Second*2)
	// if err != nil {
	// 	// printTime("yield return6 fail,err:", err, data2)
	// } else {
	// 	// printTime("yield return6:", data2.(string))
	// }

	strs := []string{"a", "b", "c"}
	for _, v := range strs {
		_, _ = co.Yield(DoSomething(v, time.Second))
		// printTime("yield return1:", data)
	}
	return nil
}

func test(task *coroutine.Task) interface{} {
	_, _ = task.Yield(func() (interface{}, error) {
		return nil, nil
	})
	sum := 0
	for i := 0; i < 1000000; i++ {
		sum += i
	}
	_, _ = task.Yield(func() (interface{}, error) {
		return nil, nil
	})
	for i := 0; i < 1000000; i++ {
		sum += i
	}
	return sum
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// ... rest of the program ...

	printTime("start")
	cg := coroutine.New(10000)
	for i := 0; i < 5000; i++ {
		cg.Add(test)
	}
	printTime("init finish")

	for {
		cg.Run()
		// printTime("main run finish")
		if cg.Len() == 0 {
			// printTime("test finish")
			break
		}
		// time.Sleep(time.Millisecond * 20)
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

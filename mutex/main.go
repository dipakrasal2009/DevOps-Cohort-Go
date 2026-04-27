package main

import (
	"fmt"
	"sync"
)

var counter int = 0

func main() {
	fmt.Println("welcome to mutex")
	var mutex sync.Mutex
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// best practice is to only lock the critical section of the code
			mutex.Lock()
			defer mutex.Unlock()
			counter++
		}()
	}

	wg.Wait()

	fmt.Println("counter:", counter)
}

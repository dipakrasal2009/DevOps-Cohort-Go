package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	Name   string
	Status int
}

func checkService(name, url string, ch chan Result, wg *sync.WaitGroup) {
	resp, err := http.Get(url)
	if err != nil {
		ch <- Result{Name: name, Status: 500}
		return
	}
	resp.Body.Close()
	wg.Done()
	ch <- Result{Name: name, Status: resp.StatusCode}
}

func main() {
	// services := []string{"api", "worker", "cron"}
	start := time.Now()

	totalServices := 100000
	channel := make(chan Result, totalServices)

	var wg sync.WaitGroup

	for i := 0; i < totalServices; i++ {
		wg.Add(1)

		servicename := fmt.Sprintf("service-%d", i)
		url := "https://example.com"
		fmt.Println("url: ", url)
		go checkService(servicename, url, channel, &wg)
	}
	wg.Wait()
	close(channel)

	for result := range channel {
		fmt.Println(result)
		// notify user that the service is ready or not
	}

	fmt.Println("Time taken:", time.Since(start))
}

/*
   for example with number
   for i := 0; i < 10; i++ {
   }

   array for example
   arr := []int{1, 2, 3, 4, 5}
   for _, v := range arr {
       fmt.Println(v)
   }

   map for example
   m := map[string]int{"a": 1, "b": 2, "c": 3}
   for k, v := range m {
       fmt.Println(k, v)
   }


*/

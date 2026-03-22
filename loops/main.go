package main

import (
	"fmt"
)

func main(){
	fmt.Println("hello, loop")

	//classic loop
	for i := 0; i<=5 ; i++ {
		fmt.Println(i)
	}

	var servers []string
	servers = []string{"server-01","server-02","server-03"}

	fmt.Println("using classic loop")

	for i := 0; i < len(servers); i++ {
		fmt.Println("index",i,"server:",servers[i])
	}

	fmt.Println("Using range:")

	for i,s := range servers {
                fmt.Println("[%d] => %s \n",i,s)

        }
	for _,s := range servers {
                fmt.Println(s)

        }

	fmt.Println("using fmt style")

	retries := 5

	for retries < 5 {
		if retries == 3 {
			retries++
			continue
		}
		fmt.Println(retries)
		retries++
	}
}


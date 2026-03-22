package main

import "fmt"

func main(){
	port := 1025
	if port > 1024 {
		fmt.Println("valid user-space port")
	} else {
		fmt.Println("Requires root privileges")
	}
}

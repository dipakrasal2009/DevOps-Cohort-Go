package main

import "fmt"

func main() {
	defer fmt.Println("hello")
	defer fmt.Println("hello 1")
	defer fmt.Println("hello 2")
	defer fmt.Println("hello 3")
	defer fmt.Println("hello 4")
	defer fmt.Println("hello 5")
	defer fmt.Println("hello 6")
	defer fmt.Println("hello 7")
	defer fmt.Println("hello 8")
	defer fmt.Println("hello 9")
	fmt.Println("world")

	fmt.Println("end of main")
}

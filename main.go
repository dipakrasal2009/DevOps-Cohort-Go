package main

import "fmt"

func main(){

        // way 1: short declaration 
        // create and assign value

        name := "api-server"
        port := 8080
        isHealthy := true


	// way 2: Explicit type

        var timeout int = 30
    

        // way 3: without any value
        var retries int  // 0
        var label string // ""
 	var active bool  // false



	fmt.Println(name, port, isHealthy ,timeout, retries, label, active)
}

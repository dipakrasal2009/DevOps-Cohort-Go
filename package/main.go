package main

import (
  "fmt"
  "os"
  str  "strings"
 "strconv"
)


func main(){

	fmt.Println("welcome to package")
	port := 8080   // create and assign, integer type
        
        msg := fmt.Sprintf("Port: %d", port)
        fmt.Println(msg)
     
      
       // os.Getenv function returns string value

       portStr := os.Getenv("MYAPP_PORT")
       fmt.Println("Port:", portStr)


	portInt, err := strconv.Atoi(portStr)
        if err != nil {
          fmt.Println("ERROR:", err)
	  os.Exit(1)
	}

	fmt.Println("Port Int:", portInt)
	

      // host variable
      host := os.Getenv("MYAPP_HOST")
      fmt.Println("Host:", host)

      
     // accept our service name
     serviceName := os.Getenv("MYAPP_NAME")
     fmt.Println("Service Name:", serviceName)

    
    // strings - manipulation
    fmt.Println(str.ToUpper(serviceName))   

}

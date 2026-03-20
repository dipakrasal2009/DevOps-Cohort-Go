package main


import "fmt"


// define service struct

type Service struct {
  Name string
  Port int
  Healthy bool
}

func main(){
  fmt.Println("hello, structs")


  httpService := Service{ Name: "gateway", Port: 8080, Healthy: true}

  fmt.Println(httpService)

  fmt.Printf("%+v\n", httpService)
}

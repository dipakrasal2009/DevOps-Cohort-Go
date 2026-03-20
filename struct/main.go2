package main


import "fmt"


// define service struct

type Service struct {
  Name string
  Port int
  Healthy bool
}

func printStatus(s Service) {

  status := "Healthy"
  
  if s.Healthy == false {
    status = "Unhealthy"
  }

 fmt.Printf("Name: %s  |  Port: %d   |  %s   \n", s.Name, s.Port, status)
}

func main(){
  fmt.Println("hello, structs")

  services := []Service{
    {Name: "gateway",  Port: 8080  , Healthy: true },
    {Name: "postgres",  Port: 5432 , Healthy: false },
    {Name: "frontend",  Port:  443 , Healthy:  true },
  }
  
  
  for _, svc := range services {
    printStatus(svc)
  }  


}

package main

import (
	"fmt"
	"time"
)

type Checker interface {
	Check()
}

type HttpService struct {
	Name    string
	URL     string
	Healthy bool
}

func (h HttpService) Check() {

	// consider we have a logic that checks if service is healthy or not using h.URL

	fmt.Println("checking if ", h.URL, "is healthy or not")
	time.Sleep(3 * time.Second)
	h.Healthy = true

	fmt.Printf("HTTP => %s, %s, %v\n", h.Name, h.URL, h.Healthy)
}

type DBService struct {
	Name    string
	Host    string
	Port    int
	Healthy bool
}

func (db DBService) Check() {

	fmt.Println("checking if ", db.Host, "is healthy or not")
	time.Sleep(3 * time.Second)

	db.Healthy = true

	fmt.Printf("DB => %s, %s:%d, %v\n", db.Name, db.Host, db.Port, db.Healthy)
}

type TCPService struct {
	Name    string
	Addr    string
	Healthy bool
}

func (t TCPService) Check() {
	fmt.Println("checking if ", t.Addr, "is healthy or not")
	time.Sleep(3 * time.Second)

	t.Healthy = true

	fmt.Printf("TCP: %s %s %v\n", t.Name, t.Addr, t.Healthy)
}

type UDPService struct {
	Name    string
	Addr    string
	Healthy bool
}

func (u UDPService) Check(){
}

func main() {

	fmt.Println("welcome to interfaces")

	services := []Checker{
		HttpService{Name: "example.com", URL: "https://example.com", Healthy: false},
		HttpService{Name: "example.org", URL: "https://example.org", Healthy: false},
		DBService{Name: "postgre", Host: "localhost", Port: 5432, Healthy: false},
		TCPService{Name: "kafka", Addr: "192.168.10.65:8765", Healthy: false},
		UDPService{Name: "udp", Addr: "192.168.10.65:8765", Healthy: false},
	}

	for _, svc := range services {
		svc.Check()
	}

}

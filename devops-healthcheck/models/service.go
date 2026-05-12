package models

import (
	"fmt"
	"net/http"
)

// Service struct definition
type Service struct {
	Timestamp   string
	Name        string
	Port        int
	URL         string
	Healthy     bool
	Description string
}

// NewService is a constructor for Service.
func NewService(name string, url string) Service {
	return Service{Name: name, URL: url}
}

// CheckHealth performs an HTTP GET to the service URL and returns true if status is 200.
func (s Service) CheckHealth() bool {
	resp, err := http.Get(s.URL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	fmt.Println("resp.StatusCode: ", resp.StatusCode)
	return resp.StatusCode == 200
}

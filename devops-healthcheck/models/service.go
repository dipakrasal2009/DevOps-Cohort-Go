package models

import (
	"fmt"
	"net/http"
)

type Service struct {
	ID        int    `json:"ID"`
	Name      string `json:"Name"`
	URL       string `json:"URL"`
	Healthy   bool   `json:"Healthy"`
	Timestamp string `json:"Timestamp"`
}

func NewService(name string, url string) Service {
	return Service{Name: name, URL: url}
}

func (s Service) CheckHealth() bool {
	resp, err := http.Get(s.URL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	fmt.Println("resp.StatusCode: ", resp.StatusCode)
	return resp.StatusCode == 200
}

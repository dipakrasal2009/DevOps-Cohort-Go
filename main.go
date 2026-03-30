package main

import (
	"fmt"

  "github.com/dipakrasal2009/DevOps-Cohort-Go/colorprint"
  "github.com/dipakrasal2009/DevOps-Cohort-Go/models")

func main() {

	fmt.Println("🚀 DevOps Health Check System")

	services := []models.Service{
		{Name: "gateway", Port: 8080, Healthy: true},
		{Name: "postgres", Port: 5432, Healthy: false},
		{Name: "frontend", Port: 443, Healthy: true},
	}

	for _, svc := range services {
		colorprint.PrintStatus(svc)
	}
}

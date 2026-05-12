package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dipakrasal2009/DevOps-Cohort-Go/devops-healthcheck/models"
)

// services acts as an in-memory store
var services map[string]models.Service

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello world")
}

func healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var service models.Service
	err = json.Unmarshal(body, &service)
	if err != nil {
		http.Error(w, "Error unmarshalling request body", http.StatusInternalServerError)
		return
	}

	service.Healthy = service.CheckHealth()
	service.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	services[service.Name] = service

	json.NewEncoder(w).Encode(service)
}

func getAllServicesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fmt.Println("total services: ", len(services))
	json.NewEncoder(w).Encode(services)
}

func runAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	for key, value := range services {
		value.Healthy = value.CheckHealth()
		value.Timestamp = time.Now().Format("2006-01-02 15:04:05")
		services[key] = value
	}

	json.NewEncoder(w).Encode(services)
}

func main() {
	fmt.Println("🚀 DevOps Health Check System")
	services = make(map[string]models.Service)

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/healthcheck", healthcheckHandler)
	http.HandleFunc("/services", getAllServicesHandler)
	http.HandleFunc("/runall", runAll)

	fmt.Println("Server running on localhost:8080")
	http.ListenAndServe(":8080", nil)
}

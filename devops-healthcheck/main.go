package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dipakrasal2009/DevOps-Cohort-Go/devops-healthcheck/db"
	"github.com/dipakrasal2009/DevOps-Cohort-Go/devops-healthcheck/models"
)

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

	// Upsert into DB
	_, err = db.DB.Exec(`
		INSERT INTO services (name, url, healthy, timestamp)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO UPDATE
		SET url=$2, healthy=$3, timestamp=$4`,
		service.Name, service.URL, service.Healthy, service.Timestamp,
	)
	if err != nil {
		fmt.Println("❌ DB Insert Error:", err)
		http.Error(w, "Error saving to DB: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("✅ Service saved:", service.Name, "| Healthy:", service.Healthy)
	json.NewEncoder(w).Encode(service)
}

func getAllServicesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Query all services from DB

	rows, err := db.DB.Query(`SELECT id, name, url, healthy, timestamp FROM services`)
	if err != nil {
		http.Error(w, "Error querying DB: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	services := []models.Service{}
	for rows.Next() {
		var s models.Service
		rows.Scan(&s.ID, &s.Name, &s.URL, &s.Healthy, &s.Timestamp)
		services = append(services, s)
	}

	json.NewEncoder(w).Encode(services)
}

func runAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.DB.Query(`SELECT id, name, url, healthy, timestamp FROM services`)
	if err != nil {
		http.Error(w, "Error querying DB: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var s models.Service
		rows.Scan(&s.ID, &s.Name, &s.URL, &s.Healthy, &s.Timestamp)
		services = append(services, s)
	}

	for i, s := range services {
		services[i].Healthy = s.CheckHealth()
		services[i].Timestamp = time.Now().Format("2006-01-02 15:04:05")

		db.DB.Exec(`
			UPDATE services SET healthy=$1, timestamp=$2 WHERE name=$3`,
			services[i].Healthy, services[i].Timestamp, s.Name,
		)
	}

	json.NewEncoder(w).Encode(services)
}

func main() {
	fmt.Println("🚀 DevOps Health Check System")

	db.Init()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/healthcheck", healthcheckHandler)
	http.HandleFunc("/services", getAllServicesHandler)
	http.HandleFunc("/runall", runAll)

	fmt.Println("Server running on localhost:8080")
	http.ListenAndServe(":8080", nil)
}

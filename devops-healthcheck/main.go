package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dipakrasal2009/DevOps-Cohort-Go/devops-healthcheck/db"
	"github.com/dipakrasal2009/DevOps-Cohort-Go/devops-healthcheck/models"
)

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// ── CREATE / health check ─────────────────────────────────────────────────────
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
	if err = json.Unmarshal(body, &service); err != nil {
		http.Error(w, "Error unmarshalling request body", http.StatusInternalServerError)
		return
	}

	service.Healthy = service.CheckHealth()
	service.Timestamp = time.Now().Format("2006-01-02 15:04:05")

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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
}

// ── READ ALL ──────────────────────────────────────────────────────────────────
func getAllServicesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.DB.Query(`SELECT id, name, url, healthy, timestamp FROM services ORDER BY id`)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// ── READ ONE ──────────────────────────────────────────────────────────────────
func getServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// /services/{name}
	name := strings.TrimPrefix(r.URL.Path, "/services/")
	if name == "" {
		http.Error(w, "Service name required", http.StatusBadRequest)
		return
	}

	var s models.Service
	err := db.DB.QueryRow(
		`SELECT id, name, url, healthy, timestamp FROM services WHERE name=$1`, name,
	).Scan(&s.ID, &s.Name, &s.URL, &s.Healthy, &s.Timestamp)
	if err != nil {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

// ── UPDATE ────────────────────────────────────────────────────────────────────
func updateServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// /services/{name}
	name := strings.TrimPrefix(r.URL.Path, "/services/")
	if name == "" {
		http.Error(w, "Service name required", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var service models.Service
	if err = json.Unmarshal(body, &service); err != nil {
		http.Error(w, "Error unmarshalling request body", http.StatusInternalServerError)
		return
	}

	// Re-check health with new URL
	service.Healthy = service.CheckHealth()
	service.Timestamp = time.Now().Format("2006-01-02 15:04:05")

	result, err := db.DB.Exec(`
		UPDATE services SET name=$1, url=$2, healthy=$3, timestamp=$4
		WHERE name=$5`,
		service.Name, service.URL, service.Healthy, service.Timestamp, name,
	)
	if err != nil {
		fmt.Println("❌ DB Update Error:", err)
		http.Error(w, "Error updating service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	fmt.Println("✅ Service updated:", service.Name)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
}

// ── DELETE ────────────────────────────────────────────────────────────────────
func deleteServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// /services/{name}
	name := strings.TrimPrefix(r.URL.Path, "/services/")
	if name == "" {
		http.Error(w, "Service name required", http.StatusBadRequest)
		return
	}

	result, err := db.DB.Exec(`DELETE FROM services WHERE name=$1`, name)
	if err != nil {
		fmt.Println("❌ DB Delete Error:", err)
		http.Error(w, "Error deleting service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	fmt.Println("✅ Service deleted:", name)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Service deleted successfully", "name": name})
}

// ── RUN ALL ───────────────────────────────────────────────────────────────────
func runAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.DB.Query(`SELECT id, name, url, healthy, timestamp FROM services ORDER BY id`)
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
		db.DB.Exec(`UPDATE services SET healthy=$1, timestamp=$2 WHERE name=$3`,
			services[i].Healthy, services[i].Timestamp, s.Name)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// ── HOME ──────────────────────────────────────────────────────────────────────
func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "🚀 DevOps Health Check API is running")
}

// ── ROUTER for /services and /services/{name} ─────────────────────────────────
func servicesRouter(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// /services — list all (GET)
	if path == "/services" || path == "/services/" {
		getAllServicesHandler(w, r)
		return
	}

	// /services/{name} — single service operations
	switch r.Method {
	case http.MethodGet:
		getServiceHandler(w, r)
	case http.MethodPut:
		updateServiceHandler(w, r)
	case http.MethodDelete:
		deleteServiceHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	fmt.Println("🚀 DevOps Health Check System")

	db.Init()

	http.HandleFunc("/", corsMiddleware(homeHandler))
	http.HandleFunc("/healthcheck", corsMiddleware(healthcheckHandler))
	http.HandleFunc("/services", corsMiddleware(servicesRouter))
	http.HandleFunc("/services/", corsMiddleware(servicesRouter))
	http.HandleFunc("/runall", corsMiddleware(runAll))

	fmt.Println("Server running on localhost:8080")
	http.ListenAndServe(":8080", nil)
}

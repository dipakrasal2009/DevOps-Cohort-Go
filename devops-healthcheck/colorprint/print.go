package utils

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/dipakrasal2009/DevOps-Cohort-Go/devops-healthcheck/models"

)

// Function to print service status
func PrintStatus(s models.Service) {

	status := "Healthy"

	if !s.Healthy {
		status = "Unhealthy"
		msg := fmt.Sprintf("Name: %s | Port: %d | %s", s.Name, s.Port, status)
		color.Red(msg)
	} else {
		msg := fmt.Sprintf("Name: %s | Port: %d | %s", s.Name, s.Port, status)
		color.HiGreen(msg)
	}
}

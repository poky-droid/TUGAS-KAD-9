package controllers

import (
	"encoding/json"
	"net/http"

	"lab-borrow/data"
	"lab-borrow/models"
)

type dashboardSummary struct {
	TotalEquipment     int                `json:"total_equipment"`
	TotalBorrowers     int                `json:"total_borrowers"`
	ActiveBorrowings   int                `json:"active_borrowings"`
	ReturnedBorrowings int                `json:"returned_borrowings"`
	AvailableItems     int                `json:"available_items"`
	RecentBorrowings   []models.Borrowing `json:"recent_borrowings"`
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	data.Mu.RLock()
	defer data.Mu.RUnlock()

	var summary dashboardSummary
	summary.TotalEquipment = len(data.Equipments)
	summary.TotalBorrowers = len(data.Borrowers)
	for _, equipment := range data.Equipments {
		summary.AvailableItems += equipment.Quantity
	}
	for _, borrowing := range data.Borrowings {
		if borrowing.Status == "Borrowed" {
			summary.ActiveBorrowings++
		} else if borrowing.Status == "Returned" {
			summary.ReturnedBorrowings++
		}
	}

	start := len(data.Borrowings) - 5
	if start < 0 {
		start = 0
	}
	summary.RecentBorrowings = make([]models.Borrowing, 0, len(data.Borrowings)-start)
	for i := len(data.Borrowings) - 1; i >= start; i-- {
		summary.RecentBorrowings = append(summary.RecentBorrowings, data.Borrowings[i])
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func AvailabilityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	data.Mu.RLock()
	defer data.Mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data.Equipments)
}

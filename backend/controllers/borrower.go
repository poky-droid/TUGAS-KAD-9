package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"lab-borrow/data"
	"lab-borrow/models"
)

func BorrowerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		data.Mu.RLock()
		defer data.Mu.RUnlock()
		getBorrowers(w, r)

	case http.MethodPost:
		data.Mu.Lock()
		defer data.Mu.Unlock()
		createBorrower(w, r)

	case http.MethodPut:
		data.Mu.Lock()
		defer data.Mu.Unlock()
		updateBorrower(w, r)

	case http.MethodDelete:
		data.Mu.Lock()
		defer data.Mu.Unlock()
		deleteBorrower(w, r)

	default:
		http.Error(w, `{"error":"Method not allowed"}`,
			http.StatusMethodNotAllowed)
	}
}

// GET: Get all borrowers or get by ID
func getBorrowers(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/borrowers/")

	if id == "" || id == "/api/borrowers" {
		json.NewEncoder(w).Encode(data.Borrowers)
		return
	}

	borrowerID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, `{"error":"Invalid ID"}`, http.StatusBadRequest)
		return
	}

	for _, borrower := range data.Borrowers {
		if borrower.ID == borrowerID {
			json.NewEncoder(w).Encode(borrower)
			return
		}
	}

	http.Error(w, `{"error":"Borrower not found"}`, http.StatusNotFound)
}

// POST: Create borrower
func createBorrower(w http.ResponseWriter, r *http.Request) {
	var borrower models.Borrower

	err := json.NewDecoder(r.Body).Decode(&borrower)
	if err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if borrower.Name == "" || borrower.StudentID == "" {
		http.Error(w, `{"error":"Name and Student ID are required"}`,
			http.StatusBadRequest)
		return
	}

	borrower.ID = 1

	for _, item := range data.Borrowers {
		if item.ID >= borrower.ID {
			borrower.ID = item.ID + 1
		}
	}

	data.Borrowers = append(data.Borrowers, borrower)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(borrower)
}

// PUT: Update borrower
func updateBorrower(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/borrowers/")

	borrowerID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, `{"error":"Invalid ID"}`, http.StatusBadRequest)
		return
	}

	var updated models.Borrower

	err = json.NewDecoder(r.Body).Decode(&updated)
	if err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	for i, borrower := range data.Borrowers {
		if borrower.ID == borrowerID {
			updated.ID = borrowerID
			data.Borrowers[i] = updated

			json.NewEncoder(w).Encode(updated)
			return
		}
	}

	http.Error(w, `{"error":"Borrower not found"}`, http.StatusNotFound)
}

// DELETE: Delete borrower
func deleteBorrower(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/borrowers/")

	borrowerID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, `{"error":"Invalid ID"}`, http.StatusBadRequest)
		return
	}

	for i, borrower := range data.Borrowers {
		if borrower.ID == borrowerID {
			for _, borrowing := range data.Borrowings {
				if borrowing.BorrowerID == borrowerID && borrowing.Status == "Borrowed" {
					http.Error(w, `{"error":"Borrower has an active borrowing"}`,
						http.StatusConflict)
					return
				}
			}

			data.Borrowers = append(
				data.Borrowers[:i],
				data.Borrowers[i+1:]...,
			)

			json.NewEncoder(w).Encode(map[string]string{
				"message": "Borrower deleted successfully",
			})
			return
		}
	}

	http.Error(w, `{"error":"Borrower not found"}`, http.StatusNotFound)
}

package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"lab-borrow/data"
	"lab-borrow/models"
)

func BorrowingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		data.Mu.RLock()
		defer data.Mu.RUnlock()
		getBorrowings(w, r)
	case http.MethodPost:
		data.Mu.Lock()
		defer data.Mu.Unlock()
		createBorrowing(w, r)
	case http.MethodPut:
		data.Mu.Lock()
		defer data.Mu.Unlock()
		updateBorrowing(w, r)
	case http.MethodDelete:
		data.Mu.Lock()
		defer data.Mu.Unlock()
		deleteBorrowing(w, r)
	case http.MethodPatch:
		data.Mu.Lock()
		defer data.Mu.Unlock()
		returnBorrowing(w, r)
	default:
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func getBorrowings(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/borrowings" {
		json.NewEncoder(w).Encode(data.Borrowings)
		return
	}

	if r.URL.Path == "/api/borrowings/history" {
		borrowings := make([]models.Borrowing, 0)
		for _, borrowing := range data.Borrowings {
			if matchesFilter(r, "borrower_id", borrowing.BorrowerID) &&
				matchesFilter(r, "equipment_id", borrowing.EquipmentID) {
				borrowings = append(borrowings, borrowing)
			}
		}
		json.NewEncoder(w).Encode(borrowings)
		return
	}

	borrowingID, ok := requestBorrowingID(r)
	if !ok {
		http.Error(w, `{"error":"Invalid ID"}`, http.StatusBadRequest)
		return
	}

	for _, borrowing := range data.Borrowings {
		if borrowing.ID == borrowingID {
			json.NewEncoder(w).Encode(borrowing)
			return
		}
	}

	http.Error(w, `{"error":"Borrowing not found"}`, http.StatusNotFound)
}

func createBorrowing(w http.ResponseWriter, r *http.Request) {
	var borrowing models.Borrowing
	if err := json.NewDecoder(r.Body).Decode(&borrowing); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if !validBorrowingDates(borrowing) {
		http.Error(w, `{"error":"Valid borrow_date and return_date are required"}`,
			http.StatusBadRequest)
		return
	}

	equipmentIndex := findEquipmentIndex(borrowing.EquipmentID)
	if equipmentIndex < 0 {
		http.Error(w, `{"error":"Equipment not found"}`, http.StatusNotFound)
		return
	}
	if !borrowerExists(borrowing.BorrowerID) {
		http.Error(w, `{"error":"Borrower not found"}`, http.StatusNotFound)
		return
	}
	if data.Equipments[equipmentIndex].Quantity < 1 {
		http.Error(w, `{"error":"Equipment is unavailable"}`, http.StatusConflict)
		return
	}

	borrowing.ID = nextBorrowingID()
	borrowing.Status = "Borrowed"
	data.Equipments[equipmentIndex].Quantity--
	data.Borrowings = append(data.Borrowings, borrowing)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(borrowing)
}

func updateBorrowing(w http.ResponseWriter, r *http.Request) {
	borrowingID, ok := requestBorrowingID(r)
	if !ok {
		http.Error(w, `{"error":"Invalid ID"}`, http.StatusBadRequest)
		return
	}

	var updated models.Borrowing
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if !validBorrowingDates(updated) {
		http.Error(w, `{"error":"Valid borrow_date and return_date are required"}`,
			http.StatusBadRequest)
		return
	}

	borrowingIndex := -1
	for i, borrowing := range data.Borrowings {
		if borrowing.ID == borrowingID {
			borrowingIndex = i
			break
		}
	}
	if borrowingIndex < 0 {
		http.Error(w, `{"error":"Borrowing not found"}`, http.StatusNotFound)
		return
	}

	old := data.Borrowings[borrowingIndex]
	if updated.Status == "" {
		updated.Status = old.Status
	}
	if updated.Status != "Borrowed" && updated.Status != "Returned" {
		http.Error(w, `{"error":"Status must be Borrowed or Returned"}`,
			http.StatusBadRequest)
		return
	}
	if !borrowerExists(updated.BorrowerID) {
		http.Error(w, `{"error":"Borrower not found"}`, http.StatusNotFound)
		return
	}

	oldEquipmentIndex := findEquipmentIndex(old.EquipmentID)
	newEquipmentIndex := findEquipmentIndex(updated.EquipmentID)
	if oldEquipmentIndex < 0 || newEquipmentIndex < 0 {
		http.Error(w, `{"error":"Equipment not found"}`, http.StatusNotFound)
		return
	}

	oldActive := old.Status == "Borrowed"
	newActive := updated.Status == "Borrowed"
	available := data.Equipments[newEquipmentIndex].Quantity
	if oldActive && oldEquipmentIndex == newEquipmentIndex {
		available++
	}
	if newActive && available < 1 {
		http.Error(w, `{"error":"Equipment is unavailable"}`, http.StatusConflict)
		return
	}

	if oldActive {
		data.Equipments[oldEquipmentIndex].Quantity++
	}
	if newActive {
		data.Equipments[newEquipmentIndex].Quantity--
	}
	updated.ID = borrowingID
	data.Borrowings[borrowingIndex] = updated
	json.NewEncoder(w).Encode(updated)
}

func deleteBorrowing(w http.ResponseWriter, r *http.Request) {
	borrowingID, ok := requestBorrowingID(r)
	if !ok {
		http.Error(w, `{"error":"Invalid ID"}`, http.StatusBadRequest)
		return
	}

	for i, borrowing := range data.Borrowings {
		if borrowing.ID == borrowingID {
			if borrowing.Status == "Borrowed" {
				equipmentIndex := findEquipmentIndex(borrowing.EquipmentID)
				if equipmentIndex >= 0 {
					data.Equipments[equipmentIndex].Quantity++
				}
			}
			data.Borrowings = append(data.Borrowings[:i], data.Borrowings[i+1:]...)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Borrowing deleted successfully",
			})
			return
		}
	}

	http.Error(w, `{"error":"Borrowing not found"}`, http.StatusNotFound)
}

func returnBorrowing(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/return") {
		http.Error(w, `{"error":"Not found"}`, http.StatusNotFound)
		return
	}

	path := strings.TrimSuffix(r.URL.Path, "/return")
	borrowingID, err := strconv.Atoi(strings.TrimPrefix(path, "/api/borrowings/"))
	if err != nil || borrowingID < 1 {
		http.Error(w, `{"error":"Invalid ID"}`, http.StatusBadRequest)
		return
	}

	for i, borrowing := range data.Borrowings {
		if borrowing.ID != borrowingID {
			continue
		}
		if borrowing.Status == "Returned" {
			http.Error(w, `{"error":"Borrowing has already been returned"}`,
				http.StatusConflict)
			return
		}
		if borrowing.Status != "Borrowed" {
			http.Error(w, `{"error":"Borrowing cannot be returned"}`,
				http.StatusConflict)
			return
		}

		equipmentIndex := findEquipmentIndex(borrowing.EquipmentID)
		if equipmentIndex < 0 {
			http.Error(w, `{"error":"Equipment not found"}`, http.StatusConflict)
			return
		}
		data.Equipments[equipmentIndex].Quantity++
		data.Borrowings[i].Status = "Returned"
		json.NewEncoder(w).Encode(data.Borrowings[i])
		return
	}

	http.Error(w, `{"error":"Borrowing not found"}`, http.StatusNotFound)
}

func requestBorrowingID(r *http.Request) (int, bool) {
	id := strings.TrimPrefix(r.URL.Path, "/api/borrowings/")
	borrowingID, err := strconv.Atoi(id)
	return borrowingID, err == nil && borrowingID > 0 && !strings.Contains(id, "/")
}

func findEquipmentIndex(id int) int {
	for i, equipment := range data.Equipments {
		if equipment.ID == id {
			return i
		}
	}
	return -1
}

func borrowerExists(id int) bool {
	for _, borrower := range data.Borrowers {
		if borrower.ID == id {
			return true
		}
	}
	return false
}

func nextBorrowingID() int {
	nextID := 1
	for _, borrowing := range data.Borrowings {
		if borrowing.ID >= nextID {
			nextID = borrowing.ID + 1
		}
	}
	return nextID
}

func validBorrowingDates(borrowing models.Borrowing) bool {
	borrowDate, borrowErr := time.Parse("2006-01-02", borrowing.BorrowDate)
	returnDate, returnErr := time.Parse("2006-01-02", borrowing.ReturnDate)
	return borrowErr == nil && returnErr == nil && !returnDate.Before(borrowDate)
}

func matchesFilter(r *http.Request, key string, value int) bool {
	filter := r.URL.Query().Get(key)
	if filter == "" {
		return true
	}
	parsed, err := strconv.Atoi(filter)
	return err == nil && parsed == value
}

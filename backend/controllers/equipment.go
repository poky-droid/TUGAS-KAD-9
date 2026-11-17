package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"lab-borrow/data"
	"lab-borrow/models"
)

func EquipmentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		data.Mu.RLock()
		defer data.Mu.RUnlock()
		getEquipments(w, r)

	case http.MethodPost:
		data.Mu.Lock()
		defer data.Mu.Unlock()
		createEquipment(w, r)

	case http.MethodPut:
		data.Mu.Lock()
		defer data.Mu.Unlock()
		updateEquipment(w, r)

	case http.MethodDelete:
		data.Mu.Lock()
		defer data.Mu.Unlock()
		deleteEquipment(w, r)

	default:
		http.Error(w, `{"error":"Method not allowed"}`,
			http.StatusMethodNotAllowed)
	}
}

// GET: Get all equipment or get by ID
func getEquipments(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/equipment/")

	if id == "" || id == "/api/equipment" {
		json.NewEncoder(w).Encode(data.Equipments)
		return
	}

	equipmentID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, `{"error":"Invalid ID"}`, http.StatusBadRequest)
		return
	}

	for _, equipment := range data.Equipments {
		if equipment.ID == equipmentID {
			json.NewEncoder(w).Encode(equipment)
			return
		}
	}

	http.Error(w, `{"error":"Equipment not found"}`, http.StatusNotFound)
}

// POST: Create equipment
func createEquipment(w http.ResponseWriter, r *http.Request) {
	var equipment models.Equipment

	err := json.NewDecoder(r.Body).Decode(&equipment)
	if err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if equipment.Name == "" || equipment.Quantity < 0 {
		http.Error(w, `{"error":"Invalid equipment data"}`,
			http.StatusBadRequest)
		return
	}

	equipment.ID = 1
	for _, item := range data.Equipments {
		if item.ID >= equipment.ID {
			equipment.ID = item.ID + 1
		}
	}

	data.Equipments = append(data.Equipments, equipment)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(equipment)
}

// PUT: Update equipment
func updateEquipment(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/equipment/")

	equipmentID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, `{"error":"Invalid ID"}`, http.StatusBadRequest)
		return
	}

	var updated models.Equipment

	err = json.NewDecoder(r.Body).Decode(&updated)
	if err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	for i, equipment := range data.Equipments {
		if equipment.ID == equipmentID {
			updated.ID = equipmentID
			data.Equipments[i] = updated

			json.NewEncoder(w).Encode(updated)
			return
		}
	}

	http.Error(w, `{"error":"Equipment not found"}`, http.StatusNotFound)
}

// DELETE: Delete equipment
func deleteEquipment(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/equipment/")

	equipmentID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, `{"error":"Invalid ID"}`, http.StatusBadRequest)
		return
	}

	for i, equipment := range data.Equipments {
		if equipment.ID == equipmentID {
			for _, borrowing := range data.Borrowings {
				if borrowing.EquipmentID == equipmentID && borrowing.Status == "Borrowed" {
					http.Error(w, `{"error":"Equipment has an active borrowing"}`,
						http.StatusConflict)
					return
				}
			}

			data.Equipments = append(
				data.Equipments[:i],
				data.Equipments[i+1:]...,
			)

			json.NewEncoder(w).Encode(map[string]string{
				"message": "Equipment deleted successfully",
			})
			return
		}
	}

	http.Error(w, `{"error":"Equipment not found"}`, http.StatusNotFound)
}

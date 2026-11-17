package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"lab-borrow/data"
	"lab-borrow/models"
)

func resetBorrowingData(t *testing.T) {
	t.Helper()
	data.Mu.Lock()
	oldEquipments := data.Equipments
	oldBorrowers := data.Borrowers
	oldBorrowings := data.Borrowings
	data.Equipments = []models.Equipment{{ID: 1, Name: "Meter", Quantity: 2}}
	data.Borrowers = []models.Borrower{{ID: 1, Name: "Student", StudentID: "S1"}}
	data.Borrowings = []models.Borrowing{}
	data.Mu.Unlock()

	t.Cleanup(func() {
		data.Mu.Lock()
		data.Equipments = oldEquipments
		data.Borrowers = oldBorrowers
		data.Borrowings = oldBorrowings
		data.Mu.Unlock()
	})
}

func performBorrowingRequest(method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	response := httptest.NewRecorder()
	BorrowingHandler(response, request)
	return response
}

func TestCreateBorrowingDecreasesEquipmentQuantity(t *testing.T) {
	resetBorrowingData(t)

	response := performBorrowingRequest(http.MethodPost, "/api/borrowings",
		`{"equipment_id":1,"borrower_id":1,"borrow_date":"2026-11-17","return_date":"2026-11-20"}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body)
	}
	if data.Equipments[0].Quantity != 1 {
		t.Fatalf("expected quantity 1 after borrowing, got %d", data.Equipments[0].Quantity)
	}
	if len(data.Borrowings) != 1 || data.Borrowings[0].Status != "Borrowed" {
		t.Fatalf("expected one active borrowing, got %#v", data.Borrowings)
	}
}

func TestReturnBorrowingRestoresQuantityOnlyOnce(t *testing.T) {
	resetBorrowingData(t)
	created := performBorrowingRequest(http.MethodPost, "/api/borrowings",
		`{"equipment_id":1,"borrower_id":1,"borrow_date":"2026-11-17","return_date":"2026-11-20"}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("could not create setup borrowing: %s", created.Body)
	}

	returned := performBorrowingRequest(http.MethodPatch, "/api/borrowings/1/return", "")
	if returned.Code != http.StatusOK {
		t.Fatalf("expected return status %d, got %d: %s", http.StatusOK, returned.Code, returned.Body)
	}
	if data.Equipments[0].Quantity != 2 || data.Borrowings[0].Status != "Returned" {
		t.Fatalf("expected returned status and restored quantity, got %#v and %d",
			data.Borrowings[0], data.Equipments[0].Quantity)
	}

	returnedAgain := performBorrowingRequest(http.MethodPatch, "/api/borrowings/1/return", "")
	if returnedAgain.Code != http.StatusConflict {
		t.Fatalf("expected duplicate return status %d, got %d", http.StatusConflict, returnedAgain.Code)
	}
	if data.Equipments[0].Quantity != 2 {
		t.Fatalf("duplicate return changed quantity to %d", data.Equipments[0].Quantity)
	}
}

func TestDeleteActiveBorrowingRestoresQuantity(t *testing.T) {
	resetBorrowingData(t)
	created := performBorrowingRequest(http.MethodPost, "/api/borrowings",
		`{"equipment_id":1,"borrower_id":1,"borrow_date":"2026-11-17","return_date":"2026-11-20"}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("could not create setup borrowing: %s", created.Body)
	}

	deleted := performBorrowingRequest(http.MethodDelete, "/api/borrowings/1", "")
	if deleted.Code != http.StatusOK {
		t.Fatalf("expected delete status %d, got %d: %s", http.StatusOK, deleted.Code, deleted.Body)
	}
	if data.Equipments[0].Quantity != 2 || len(data.Borrowings) != 0 {
		t.Fatalf("expected deleted borrowing and restored quantity, got %#v and %d",
			data.Borrowings, data.Equipments[0].Quantity)
	}
}

func TestCreateBorrowingRejectsUnavailableEquipment(t *testing.T) {
	resetBorrowingData(t)
	data.Equipments[0].Quantity = 0

	response := performBorrowingRequest(http.MethodPost, "/api/borrowings",
		`{"equipment_id":1,"borrower_id":1,"borrow_date":"2026-11-17","return_date":"2026-11-20"}`)
	if response.Code != http.StatusConflict {
		t.Fatalf("expected unavailable equipment status %d, got %d", http.StatusConflict, response.Code)
	}
	if len(data.Borrowings) != 0 {
		t.Fatalf("unavailable equipment created a borrowing: %#v", data.Borrowings)
	}
}

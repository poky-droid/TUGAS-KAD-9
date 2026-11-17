package data

import (
	"sync"

	"lab-borrow/models"
)

var Mu sync.RWMutex

var Equipments = []models.Equipment{
	{
		ID:        1,
		Name:      "Multimeter",
		Category:  "Electronics",
		Quantity:  4,
		Condition: "Good",
	},
	{
		ID:        2,
		Name:      "Oscilloscope",
		Category:  "Electronics",
		Quantity:  3,
		Condition: "Good",
	},
}

var Borrowers = []models.Borrower{
	{
		ID:        1,
		Name:      "Ahmad",
		StudentID: "123456",
		Email:     "ahmad@example.com",
		Phone:     "08123456789",
	},
	{
		ID:        2,
		Name:      "Budi",
		StudentID: "123457",
		Email:     "budi@example.com",
		Phone:     "08123456780",
	},
}

var Borrowings = []models.Borrowing{
	{
		ID:          1,
		EquipmentID: 1,
		BorrowerID:  1,
		BorrowDate:  "2026-11-17",
		ReturnDate:  "2026-11-20",
		Status:      "Borrowed",
	},
}

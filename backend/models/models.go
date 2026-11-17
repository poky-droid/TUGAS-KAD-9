package models

type Equipment struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Quantity  int    `json:"quantity"`
	Condition string `json:"condition"`
}

type Borrower struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	StudentID string `json:"student_id"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

type Borrowing struct {
	ID          int    `json:"id"`
	EquipmentID int    `json:"equipment_id"`
	BorrowerID  int    `json:"borrower_id"`
	BorrowDate  string `json:"borrow_date"`
	ReturnDate  string `json:"return_date"`
	Status      string `json:"status"`
}

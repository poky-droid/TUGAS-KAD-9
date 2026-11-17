package routes

import (
	"net/http"

	"lab-borrow/controllers"
)

func SetupRoutes() {
	// Equipment Routes
	http.HandleFunc("/api/equipment", controllers.EquipmentHandler)
	http.HandleFunc("/api/equipment/", controllers.EquipmentHandler)

	// Borrower Routes
	http.HandleFunc("/api/borrowers", controllers.BorrowerHandler)
	http.HandleFunc("/api/borrowers/", controllers.BorrowerHandler)

	// Borrowing Routes
	http.HandleFunc("/api/borrowings", controllers.BorrowingHandler)
	http.HandleFunc("/api/borrowings/", controllers.BorrowingHandler)

	// Summary Routes
	http.HandleFunc("/api/dashboard", controllers.DashboardHandler)
	http.HandleFunc("/api/availability", controllers.AvailabilityHandler)

	http.Handle("/", http.FileServer(http.Dir("../frontend")))
}

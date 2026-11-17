package main

import (
	"fmt"
	"log"
	"net/http"

	"lab-borrow/routes"
)

func main() {
	routes.SetupRoutes()

	fmt.Println("LabBorrow API running on http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}

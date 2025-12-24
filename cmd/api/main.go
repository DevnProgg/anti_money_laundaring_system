package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"anti-money-laundering-system/internal/config"
	"anti-money-laundering-system/internal/handlers"
)

func main() {
	fmt.Println("Starting Anti Money Laundering System API...")

	rules, err := config.LoadRules("rules.json")
	if err != nil {
		log.Fatalf("Failed to load rules: %v", err)
	}
	fmt.Printf("Loaded %d rules\n", len(rules))

	// Placeholder for database connection
	var db *sql.DB

	http.HandleFunc("/transactions", handlers.TransactionHandler(db))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

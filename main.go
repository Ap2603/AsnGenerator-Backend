package main

import (
	"AsnGenerator-Backend/db"
	"AsnGenerator-Backend/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	log.Println("Raw DATABASE_URL:", dbURL)

	// Initialize database connection
	db.InitDB(dbURL)
	db.CreateSchema()

	// Register handlers
	http.HandleFunc("/api/register", handlers.AuthMiddleware(handlers.AdminMiddleware(handlers.RegisterHandler)))
	http.HandleFunc("/api/login", handlers.LoginHandler)

	http.HandleFunc("/api/importBadger", handlers.ImportBadgerHandler)
	http.HandleFunc("/api/importPO", handlers.AuthMiddleware(handlers.AdminMiddleware(handlers.ImportPOHandler)))
	http.HandleFunc("/api/exportASN", handlers.AuthMiddleware(handlers.AdminMiddleware(handlers.ExportASNHandler)))
	http.HandleFunc("/api/queryBarcode", handlers.AuthMiddleware(handlers.BarcodeHandler))
	http.HandleFunc("/api/getShipmentIDs", handlers.AuthMiddleware(handlers.GetShipmentIDsHandler))
	http.HandleFunc("/api/getPONumbers", handlers.AuthMiddleware(handlers.GetPONumbers))
	http.HandleFunc("/api/viewPO", handlers.AuthMiddleware(handlers.AdminMiddleware(handlers.ViewPOHandler)))
	http.HandleFunc("/api/addShipmentID", handlers.AuthMiddleware(handlers.AdminMiddleware(handlers.AddShipmentIDHandler)))
    http.HandleFunc("/api/viewASNStatus", handlers.AuthMiddleware(handlers.AdminMiddleware(handlers.ViewASNHandler)))
    http.HandleFunc("/api/removeASNEntry", handlers.AuthMiddleware(handlers.AdminMiddleware(handlers.RemoveASNEntryHandler)))


	log.Println("Starting server on :7000...")
	log.Fatal(http.ListenAndServe(":7000", nil))
}

// func insertAdminUser() {
//     username := "admin"
//     password := "Fivestar@2755!" // Change this to the desired admin password
//     hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
//     if err != nil {
//         log.Fatalf("Error generating password hash: %v", err)
//     }

//     _, err = db.GetDB().Exec("INSERT INTO Users (username, password, role) VALUES ($1, $2, $3)", username, hashedPassword, "admin")
//     if err != nil {
//         log.Printf("Error inserting admin user: %v", err)
//     } else {
//         log.Println("Admin user inserted successfully")
//     }
// }

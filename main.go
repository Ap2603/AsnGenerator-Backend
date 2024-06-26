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
    http.HandleFunc("/api/importBadger", handlers.ImportBadgerHandler)
    http.HandleFunc("/api/importPO", handlers.ImportPOHandler)
    http.HandleFunc("/api/queryBarcode", handlers.BarcodeHandler)
    http.HandleFunc("/api/exportASN", handlers.ExportASNHandler)
    // http.HandleFunc("/api/addShipmentID", handlers.AddShipmentIDHandler)
    // http.HandleFunc("/api/viewPO", handlers.ViewPOHandler)
    // http.HandleFunc("/api/resetTables", handlers.ResetTablesHandler)

    log.Println("Starting server on :8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}


// func rootHandler(w http.ResponseWriter, r *http.Request) {
//     fmt.Fprintln(w, "Welcome to the Inventory App")
// }

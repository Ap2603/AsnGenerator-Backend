package main 

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"

    _ "github.com/mattn/go-sqlite3"
)

var db *sql.DB


func main() {
    // Open the database
    var err error
    db, err = sql.Open("sqlite3", "./inventory.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create tables
    err = createTables()
    if err != nil {
        log.Fatal(err)
    }

    // Set up HTTP routes
    http.HandleFunc("/api/login", loginHandler)
    http.HandleFunc("/api/register", registerHandler)
    http.HandleFunc("/api/importBadger", importBadgerHandler)
    http.HandleFunc("/api/importPO", importPOHandler)
    http.HandleFunc("/api/queryBarcode", queryBarcodeHandler)
    http.HandleFunc("/api/exportAsn", exportAsnHandler)
    http.HandleFunc("/api/resetTables", resetTablesHandler)

    // Start the server
    fmt.Println("Starting server on :8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func createTables() error {
    schema, err := os.ReadFile("schema.sql")
    if err != nil {
        return err
    }

    _, err = db.Exec(string(schema))
    return err
}

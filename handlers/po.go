package handlers

import (
    "AsnGenerator-Backend/db"
    "encoding/json"
    "log"
    "net/http"
    "database/sql"
)

func GetPONumbers(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("query")

    var rows *sql.Rows
    var err error

    if query == "" {
        rows, err = db.GetDB().Query("SELECT PO_Number FROM PO")
    } else {
        rows, err = db.GetDB().Query("SELECT PO_Number FROM PO WHERE PO_Number LIKE $1", "%"+query+"%")
    }

    if err != nil {
        log.Println("Error fetching PO numbers:", err)
        http.Error(w, "Error fetching PO numbers", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var poNumbers []string
    for rows.Next() {
        var poNumber string
        if err := rows.Scan(&poNumber); err != nil {
            log.Println("Error scanning PO number:", err)
            http.Error(w, "Error scanning PO number", http.StatusInternalServerError)
            return
        }
        poNumbers = append(poNumbers, poNumber)
    }

    if err := rows.Err(); err != nil {
        log.Println("Error iterating over PO numbers:", err)
        http.Error(w, "Error iterating over PO numbers", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(poNumbers)
}

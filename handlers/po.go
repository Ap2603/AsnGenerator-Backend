package handlers

import (
    "AsnGenerator-Backend/db"
    "encoding/json"
    "net/http"
)

func GetPONumbers(w http.ResponseWriter, r *http.Request) {
    rows, err := db.GetDB().Query("SELECT PO_Number FROM PO")
    if err != nil {
        http.Error(w, "Error fetching PO numbers", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var poNumbers []string
    for rows.Next() {
        var poNumber string
        if err := rows.Scan(&poNumber); err != nil {
            http.Error(w, "Error scanning PO number", http.StatusInternalServerError)
            return
        }
        poNumbers = append(poNumbers, poNumber)
    }

    if err := rows.Err(); err != nil {
        http.Error(w, "Error iterating over PO numbers", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(poNumbers)
}

package handlers

import (
	"AsnGenerator-Backend/db"
	"AsnGenerator-Backend/structs"
	"encoding/json"
	"log"
	"net/http"
)

func ViewPOHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    poNumber := r.URL.Query().Get("poNumber")
    if poNumber == "" {
        http.Error(w, "Missing PO number", http.StatusBadRequest)
        return
    }

    showFinished := r.URL.Query().Get("showFinished")
    hideFinished := r.URL.Query().Get("hideFinished")

    query := `
        SELECT Line_Number, Item_Number, Style, Colour_Size, Cost, Pcs, Total, Ex_Fac_Date
        FROM ItemsOrdered
        WHERE PO_Number = $1
    `

    if showFinished == "true" {
        query += " AND Pcs <= 0"
    } else if hideFinished == "true" {
        query += " AND Pcs > 0"
    }

    query += " ORDER BY Line_Number ASC"

    rows, err := db.GetDB().Query(query, poNumber)
    if err != nil {
        log.Println("Error querying ItemsOrdered table:", err)
        http.Error(w, "Error querying ItemsOrdered table", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var results []structs.POresults

    for rows.Next() {
        var entry structs.POresults
        if err := rows.Scan(&entry.LineNumber, &entry.ItemNumber, &entry.Style, &entry.ColourSize, &entry.Cost, &entry.Pcs, &entry.Total, &entry.ExFacDate); err != nil {
            log.Println("Error scanning row:", err)
            http.Error(w, "Error scanning row", http.StatusInternalServerError)
            return
        }
        results = append(results, entry)
    }

    if err := rows.Err(); err != nil {
        log.Println("Error iterating rows:", err)
        http.Error(w, "Error iterating rows", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(results)
}

package handlers

import (
	"AsnGenerator-Backend/db"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

func GetShipmentIDsHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("query")

    var rows *sql.Rows
    var err error

    if query == "" {
        rows, err = db.GetDB().Query("SELECT ShipmentID FROM ShipmentID")
    } else {
        rows, err = db.GetDB().Query("SELECT ShipmentID FROM ShipmentID WHERE ShipmentID LIKE $1", "%"+query+"%")
    }

    if err != nil {
        log.Println("Error fetching shipment IDs:", err)
        http.Error(w, "Error fetching shipment IDs", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var shipmentIDs []string
    for rows.Next() {
        var shipmentID string
        if err := rows.Scan(&shipmentID); err != nil {
            log.Println("Error scanning shipment ID:", err)
            http.Error(w, "Error scanning shipment ID", http.StatusInternalServerError)
            return
        }
        shipmentIDs = append(shipmentIDs, shipmentID)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(shipmentIDs)
}

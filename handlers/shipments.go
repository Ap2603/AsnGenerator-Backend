package handlers

import (
    "AsnGenerator-Backend/db"
    "encoding/json"
    "log"
    "net/http"
)

func GetShipmentIDsHandler(w http.ResponseWriter, r *http.Request) {
    rows, err := db.GetDB().Query("SELECT ShipmentID FROM ShipmentID")
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

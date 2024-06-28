package handlers

import (
	"AsnGenerator-Backend/db"
	"AsnGenerator-Backend/structs"
	"encoding/json"
	"log"
	"net/http"
)



func AddShipmentIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req structs.AddShipmentIDRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("Error decoding request:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.ShipmentID == "" {
		http.Error(w, "Shipment ID is required", http.StatusBadRequest)
		return
	}

	_, err = db.GetDB().Exec(`INSERT INTO shipmentID (ShipmentID) VALUES ($1)`, req.ShipmentID)
	if err != nil {
		log.Println("Error inserting into Shipments table:", err)
		http.Error(w, "Error inserting into Shipments table", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Shipment ID added successfully"})
}

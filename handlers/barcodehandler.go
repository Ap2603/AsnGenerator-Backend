package handlers

import (
	"AsnGenerator-Backend/db"
	"AsnGenerator-Backend/structs"
	"encoding/json"
	"log"
	"net/http"
)

func BarcodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response := map[string]string{"message": "Method not allowed"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	var req structs.BarcodeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("Error decoding request:", err)
		response := map[string]string{"message": "Invalid request payload", "error": err.Error()}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("Received request: %+v", req)
	log.Printf("Received PO Number: %s", req.PONumber)
	log.Printf("Received Shipment ID: %s", req.ShipmentID)

	if len(req.GTIN) != 20 || len(req.SSCC) != 20 {
		response := map[string]string{"message": "Both GTIN and SSCC must be 20 characters long"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check if GTIN and SSCC are the same
	if req.GTIN == req.SSCC {
		response := map[string]string{"message": "GTIN and SSCC cannot be the same"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	filteredGTIN := req.GTIN[2 : len(req.GTIN)-4]
	log.Printf("Filtered GTIN: %v", filteredGTIN)

	var itemCode string
	var casePackSize int
	err = db.GetDB().QueryRow(`
		SELECT M3_Sku, M3_std_case_pack_size
		FROM Badger
		WHERE GTIN_Alias = $1
	`, filteredGTIN).Scan(&itemCode, &casePackSize)
	if err != nil {
		log.Println("Error querying Badger table:", err)
		response := map[string]string{"message": "Error querying Badger table", "error": err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("Item Code: %v, Case Pack Size: %d", itemCode, casePackSize)

	var lineNumber int
	var currentPcs int
	err = db.GetDB().QueryRow(`
		SELECT Line_Number, Pcs
		FROM ItemsOrdered
		WHERE PO_Number = $1 AND Item_Number = $2
	`, req.PONumber, itemCode).Scan(&lineNumber, &currentPcs)
	if err != nil {
		log.Printf("Error querying ItemsOrdered table for PO_Number %s and Item_Number %s: %v", req.PONumber, itemCode, err)
		response := map[string]string{"message": "Error querying ItemsOrdered table. No such item in selected PO.", "error": err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("Line Number: %d, Current PCS: %d", lineNumber, currentPcs)

	if currentPcs < casePackSize && !req.Override {
		var response map[string]string
		if req.Role == "admin" {
			response = map[string]string{"message": "Not enough PCS in ItemsOrdered to fulfill this ASN entry", "allowOverride": "true"}
		} else {
			response = map[string]string{"message": "Not enough PCS in ItemsOrdered to fulfill this ASN entry", "allowOverride": "false"}
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	filteredSSCC := req.SSCC[2:]
	log.Printf("Filtered SSCC: %v", filteredSSCC)

	tx, err := db.GetDB().Begin()
	if err != nil {
		log.Println("Error beginning transaction:", err)
		response := map[string]string{"message": "Error beginning transaction", "error": err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO ASN (SSCC, Item_Code, Case_Pack_Size, PO_Number, Line_Number, ShipmentID)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, filteredSSCC, itemCode, casePackSize, req.PONumber, lineNumber, req.ShipmentID)
	if err != nil {
		log.Println("Error inserting into ASN table:", err)
		response := map[string]string{"message": "Error inserting into ASN table", "error": err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	if currentPcs >= casePackSize || (currentPcs < casePackSize && req.Override) {
		_, err = tx.Exec(`
			UPDATE ItemsOrdered
			SET Pcs = Pcs - $1
			WHERE PO_Number = $2 AND Item_Number = $3
		`, casePackSize, req.PONumber, itemCode)
		if err != nil {
			log.Println("Error updating ItemsOrdered table:", err)
			response := map[string]string{"message": "Error updating ItemsOrdered table", "error": err.Error()}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
		response := map[string]string{"message": "Error committing transaction", "error": err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Data processed successfully"})
}

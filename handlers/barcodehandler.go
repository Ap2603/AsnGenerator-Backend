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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req structs.BarcodeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("Error decoding request:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate barcodes
	if len(req.GTIN) != 20 || len(req.SSCC) != 20 {
		http.Error(w, "Both GTIN and SSCC must be 20 characters long", http.StatusBadRequest)
		return
	}

	// Filter GTIN
	filteredGTIN := req.GTIN[2 : len(req.GTIN)-4]

	// Query Badger table
	var itemCode string
	var casePackSize int
	err = db.GetDB().QueryRow(`
		SELECT M3_Sku, M3_std_case_pack_size
		FROM Badger
		WHERE GTIN_Alias = $1
	`, filteredGTIN).Scan(&itemCode, &casePackSize)
	if err != nil {
		log.Println("Error querying Badger table:", err)
		http.Error(w, "Error querying Badger table", http.StatusInternalServerError)
		return
	}

	// Query PO tables and get Line number and current pcs
	var lineNumber int
	var currentPcs int
	err = db.GetDB().QueryRow(`
		SELECT Line_Number, Pcs
		FROM ItemsOrdered
		WHERE PO_Number = $1 AND Item_Number = $2
	`, req.PONumber, itemCode).Scan(&lineNumber, &currentPcs)
	if err != nil {
		log.Println("Error querying ItemsOrdered table:", err)
		http.Error(w, "Error querying ItemsOrdered table", http.StatusInternalServerError)
		return
	}

	// Check if subtracting casePackSize will result in negative pcs
	if currentPcs < casePackSize {
		http.Error(w, "Not enough PCS in ItemsOrdered to fulfill this ASN entry", http.StatusBadRequest)
		return
	}

	// Filter SSCC
	filteredSSCC := req.SSCC[2:]

	// Begin transaction
	tx, err := db.GetDB().Begin()
	if err != nil {
		log.Println("Error beginning transaction:", err)
		http.Error(w, "Error beginning transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert into ASN table
	_, err = tx.Exec(`
		INSERT INTO ASN (SSCC, Item_Code, Case_Pack_Size, PO_Number, Line_Number, ShipmentID)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, filteredSSCC, itemCode, casePackSize, req.PONumber, lineNumber, req.ShipmentID)
	if err != nil {
		log.Println("Error inserting into ASN table:", err)
		http.Error(w, "Error inserting into ASN table", http.StatusInternalServerError)
		return
	}

	// Update ItemsOrdered table
	_, err = tx.Exec(`
		UPDATE ItemsOrdered
		SET Pcs = Pcs - $1
		WHERE PO_Number = $2 AND Item_Number = $3
	`, casePackSize, req.PONumber, itemCode)
	if err != nil {
		log.Println("Error updating ItemsOrdered table:", err)
		http.Error(w, "Error updating ItemsOrdered table", http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Data processed successfully"})
}

package handlers

import (
	"AsnGenerator-Backend/db"
	"AsnGenerator-Backend/structs"
	"encoding/json"
	"log"
	"net/http"
)

func RemoveASNEntryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response := map[string]string{"message": "Method not allowed"}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	var req structs.EntryRemoval 


	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("Error decoding request:", err)
		response := map[string]string{"message": "Invalid request payload"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get the case pack size for the item from ASN table before deleting
	var casePackSize int
	err = db.GetDB().QueryRow(`
		SELECT Case_Pack_Size
		FROM ASN
		WHERE SSCC = $1 AND PO_Number = $2 AND Item_Code = $3 AND Line_Number = $4
	`, req.SSCC, req.PONumber, req.ItemCode, req.LineNumber).Scan(&casePackSize)
	if err != nil {
		log.Println("Error querying ASN table:", err)
		response := map[string]string{"message": "Error querying ASN table"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Begin transaction
	tx, err := db.GetDB().Begin()
	if err != nil {
		log.Println("Error beginning transaction:", err)
		response := map[string]string{"message": "Error beginning transaction"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer tx.Rollback()

	// Delete from ASN table
	_, err = tx.Exec(`
		DELETE FROM ASN
		WHERE SSCC = $1 AND PO_Number = $2 AND Item_Code = $3 AND Line_Number = $4
	`, req.SSCC, req.PONumber, req.ItemCode, req.LineNumber)
	if err != nil {
		log.Println("Error deleting from ASN table:", err)
		response := map[string]string{"message": "Error deleting from ASN table"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Update ItemsOrdered table
	_, err = tx.Exec(`
		UPDATE ItemsOrdered
		SET Pcs = Pcs + $1
		WHERE PO_Number = $2 AND Item_Number = $3 AND Line_Number = $4
	`, casePackSize, req.PONumber, req.ItemCode, req.LineNumber)
	if err != nil {
		log.Println("Error updating ItemsOrdered table:", err)
		response := map[string]string{"message": "Error updating ItemsOrdered table"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
		response := map[string]string{"message": "Error committing transaction"}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Entry removed successfully"})
}

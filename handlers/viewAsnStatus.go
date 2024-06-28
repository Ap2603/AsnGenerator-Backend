package handlers

import (
	"AsnGenerator-Backend/db"
	"AsnGenerator-Backend/structs"
	"encoding/json"
	"log"
	"net/http"
)

func ViewASNHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shipmentId := r.URL.Query().Get("shipmentId")
	if shipmentId == "" {
		http.Error(w, "Missing shipment ID", http.StatusBadRequest)
		log.Println("Missing shipment ID")
		return
	}

	log.Printf("Fetching ASN for shipment ID: %s\n", shipmentId)

	rows, err := db.GetDB().Query(`
		SELECT SSCC, Item_Code, Case_Pack_Size, PO_Number, Line_Number
		FROM ASN
		WHERE ShipmentID = $1
		ORDER BY Line_Number ASC
	`, shipmentId)
	if err != nil {
		log.Println("Error querying ASN table:", err)
		http.Error(w, "Error querying ASN table", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []structs.ASNResults

	for rows.Next() {
		var entry structs.ASNResults
		if err := rows.Scan(&entry.SSCC, &entry.ItemCode, &entry.CasePackSize, &entry.PONumber, &entry.LineNumber); err != nil {
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

	log.Printf("ASN results: %v\n", results)

	w.Header().Set("Content-Type", "application/json")
	if len(results) == 0 {
		log.Println("No ASN entries found for the given shipment ID")
		json.NewEncoder(w).Encode([]struct{}{})
	} else {
		json.NewEncoder(w).Encode(results)
	}
}

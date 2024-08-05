package handlers

import (
	"AsnGenerator-Backend/db"
	"AsnGenerator-Backend/structs"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

func AddPOentriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var items []structs.POentries
	err := json.NewDecoder(r.Body).Decode(&items)
	if err != nil {
		log.Println("Error decoding request:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	tx, err := db.GetDB().Begin()
	if err != nil {
		log.Println("Error beginning transaction:", err)
		http.Error(w, "Error beginning transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, item := range items {
		// Check if PO number exists
		var exists bool
		err := db.GetDB().QueryRow(`
			SELECT EXISTS(SELECT 1 FROM PO WHERE PO_Number = $1)
		`, item.PONumber).Scan(&exists)
		if err != nil {
			log.Println("Error checking PO number:", err)
			http.Error(w, "Error checking PO number", http.StatusInternalServerError)
			return
		}

		if !exists {
			log.Printf("PO number %s does not exist\n", item.PONumber)
			http.Error(w, "PO number does not exist", http.StatusBadRequest)
			return
		}

		_, err = tx.Exec(`
			INSERT INTO ItemsOrdered (
				PO_Number, Line_Number, Item_Number, Style, Colour_Size, Cost, Pcs, Total, Ex_Fac_Date
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, item.PONumber, item.LineNumber, item.ItemNumber, item.Style, item.ColourSize, item.Cost, item.Pcs, item.Total, item.ExFacDate)
		if err != nil {
			log.Println("Error inserting into ItemsOrdered table:", err)
			http.Error(w, "Error inserting into ItemsOrdered table", http.StatusInternalServerError)
			return
		}

		// Remove any non-numeric characters from the total value
		re := regexp.MustCompile(`[^\d\.]`)
		cleanedTotal := re.ReplaceAllString(item.Total, "")

		// Convert cleaned total to a float for addition
		newTotal, err := strconv.ParseFloat(cleanedTotal, 64)
		if err != nil {
			log.Println("Error converting total to float:", err)
			http.Error(w, "Error converting total to float", http.StatusInternalServerError)
			return
		}

		// Remove any non-numeric characters from the existing total value in PO table
		var existingTotal string
		err = tx.QueryRow(`
			SELECT Total
			FROM PO
			WHERE PO_Number = $1
		`, item.PONumber).Scan(&existingTotal)
		if err != nil {
			log.Println("Error retrieving existing total:", err)
			http.Error(w, "Error retrieving existing total", http.StatusInternalServerError)
			return
		}

		existingTotalClean := re.ReplaceAllString(existingTotal, "")

		existingTotalFloat, err := strconv.ParseFloat(existingTotalClean, 64)
		if err != nil {
			log.Println("Error converting existing total to float:", err)
			http.Error(w, "Error converting existing total to float", http.StatusInternalServerError)
			return
		}

		// Calculate the new total
		combinedTotal := existingTotalFloat + newTotal

		// Format the combined total with a $ symbol
		formattedTotal := fmt.Sprintf("$%.2f", combinedTotal)

		// Update total in PO table
		_, err = tx.Exec(`
			UPDATE PO
			SET Total = $1
			WHERE PO_Number = $2
		`, formattedTotal, item.PONumber)
		if err != nil {
			log.Println("Error updating total in PO table:", err)
			http.Error(w, "Error updating total in PO table", http.StatusInternalServerError)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Items added successfully"})
}

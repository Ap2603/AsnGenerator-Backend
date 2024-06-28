package handlers

import (
	"AsnGenerator-Backend/db"
	"bytes"
	"github.com/xuri/excelize/v2"
	"log"
	"net/http"
	"strconv"
)

func ExportASNHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shipmentID := r.URL.Query().Get("shipment_id")
	if shipmentID == "" {
		http.Error(w, "shipment_id query parameter is required", http.StatusBadRequest)
		return
	}

	log.Printf("Received request to export ASN for shipment_id: %s", shipmentID)

	// Query ASN data for the given shipment ID
	rows, err := db.GetDB().Query(`
		SELECT SSCC, Item_Code, Case_Pack_Size, PO_Number, Line_Number
		FROM ASN
		WHERE ShipmentID = $1
		ORDER BY Item_Code ASC
	`, shipmentID)
	if err != nil {
		log.Println("Error querying ASN data:", err)
		http.Error(w, "Error querying ASN data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create a new Excel file and rename the default sheet
	f := excelize.NewFile()
	sheetName := "ASN Data"
	f.SetSheetName("Sheet1", sheetName)

	// Define header and data style
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	if err != nil {
		log.Println("Error creating header style:", err)
		http.Error(w, "Error creating header style", http.StatusInternalServerError)
		return
	}

	dataStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	if err != nil {
		log.Println("Error creating data style:", err)
		http.Error(w, "Error creating data style", http.StatusInternalServerError)
		return
	}

	// Set headers and column widths
	headers := []string{"SSCC", "Item code", "Pieces per carton", "PO number", "Line number"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
		f.SetColWidth(sheetName, cell[:1], cell[:1], 20) // Set column width to 20 (double the default)
	}

	// Populate data
	rowNumber := 2
	for rows.Next() {
		var sscc, itemCode, poNumber string
		var casePackSize, lineNumber int
		if err := rows.Scan(&sscc, &itemCode, &casePackSize, &poNumber, &lineNumber); err != nil {
			log.Println("Error scanning ASN data:", err)
			http.Error(w, "Error scanning ASN data", http.StatusInternalServerError)
			return
		}
		f.SetCellValue(sheetName, "A"+strconv.Itoa(rowNumber), sscc)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(rowNumber), itemCode)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(rowNumber), casePackSize)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(rowNumber), poNumber)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(rowNumber), lineNumber)
		for col := 1; col <= 5; col++ {
			cell, _ := excelize.CoordinatesToCellName(col, rowNumber)
			f.SetCellStyle(sheetName, cell, cell, dataStyle)
		}
		rowNumber++
	}
	sheetindex, err := f.GetSheetIndex(sheetName)
	if err != nil{
		log.Println("Error finding sheet index:", err)
		http.Error(w, "Error finding sheet index", http.StatusInternalServerError)
		return
	}
	// Set the active sheet
	f.SetActiveSheet(sheetindex)

	// Write the file to a buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		log.Println("Error writing Excel file:", err)
		http.Error(w, "Error writing Excel file", http.StatusInternalServerError)
		return
	}

	// Generate the file name
	fileName := shipmentID + "-ASN.xlsx"

	// Set the headers for file download
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))

	// Write the buffer to the response
	if _, err := buf.WriteTo(w); err != nil {
		log.Println("Error sending file:", err)
		http.Error(w, "Error sending file", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully sent ASN file: %s", fileName)
}




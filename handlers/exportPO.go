package handlers

import (
	"AsnGenerator-Backend/db"
	"log"
	"net/http"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func ExportPOHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	poNumber := r.URL.Query().Get("poNumber")
	if poNumber == "" {
		http.Error(w, "Missing PO number", http.StatusBadRequest)
		log.Println("Missing PO number")
		return
	}

	log.Printf("Exporting PO for PO Number: %s\n", poNumber)

	rows, err := db.GetDB().Query(`
		SELECT Line_Number, Item_Number, Style, Colour_Size, Cost, Pcs, Total, Ex_Fac_Date
		FROM ItemsOrdered
		WHERE PO_Number = $1
		ORDER BY Line_Number ASC
	`, poNumber)
	if err != nil {
		log.Println("Error querying ItemsOrdered table:", err)
		http.Error(w, "Error querying ItemsOrdered table", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	f := excelize.NewFile()
	sheetName := "PO_" + poNumber
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")

	// Set the headers
	headers := []string{"Line Number", "Item Number", "Style", "Colour/Size", "Cost", "Pcs", "Total", "Ex Fac Date"}
	for i, header := range headers {
		cell := string(rune('A'+i)) + "1"
		f.SetCellValue(sheetName, cell, header)
	}

	// Fill in the data
	rowNumber := 2
	for rows.Next() {
		var lineNumber, pcs int
		var itemNumber, style, colourSize, cost, total, exFacDate string
		if err := rows.Scan(&lineNumber, &itemNumber, &style, &colourSize, &cost, &pcs, &total, &exFacDate); err != nil {
			log.Println("Error scanning row:", err)
			http.Error(w, "Error scanning row", http.StatusInternalServerError)
			return
		}

		f.SetCellValue(sheetName, "A"+strconv.Itoa(rowNumber), lineNumber)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(rowNumber), itemNumber)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(rowNumber), style)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(rowNumber), colourSize)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(rowNumber), cost)
		f.SetCellValue(sheetName, "F"+strconv.Itoa(rowNumber), pcs)
		f.SetCellValue(sheetName, "G"+strconv.Itoa(rowNumber), total)
		f.SetCellValue(sheetName, "H"+strconv.Itoa(rowNumber), exFacDate)

		rowNumber++
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating rows:", err)
		http.Error(w, "Error iterating rows", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=PO_"+poNumber+".xlsx")

	if err := f.Write(w); err != nil {
		log.Println("Error writing Excel file:", err)
		http.Error(w, "Error writing Excel file", http.StatusInternalServerError)
		return
	}
}

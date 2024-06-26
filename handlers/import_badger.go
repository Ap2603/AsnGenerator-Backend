package handlers

import (
    "AsnGenerator-Backend/db"
    "encoding/json"
    "net/http"
    "os"
    "io"
    "path/filepath"
    "github.com/xuri/excelize/v2"
    "log"
    "strconv"
    "strings"
)

func ImportBadgerHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Parse the multipart form
    err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
    if err != nil {
        log.Println("Error parsing form:", err)
        http.Error(w, "Error parsing form", http.StatusInternalServerError)
        return
    }

    // Get the uploaded file
    file, handler, err := r.FormFile("file")
    if err != nil {
        log.Println("Error retrieving file:", err)
        http.Error(w, "Error retrieving file", http.StatusInternalServerError)
        return
    }
    defer file.Close()

    // Save the file locally
    tempFile, err := os.Create(filepath.Join(os.TempDir(), handler.Filename))
    if err != nil {
        log.Println("Error creating temp file:", err)
        http.Error(w, "Error saving file", http.StatusInternalServerError)
        return
    }
    defer tempFile.Close()

    _, err = io.Copy(tempFile, file)
    if err != nil {
        log.Println("Error saving file:", err)
        http.Error(w, "Error saving file", http.StatusInternalServerError)
        return
    }

    // Open the Excel file
    f, err := excelize.OpenFile(tempFile.Name())
    if err != nil {
        log.Println("Error opening Excel file:", err)
        http.Error(w, "Error opening Excel file", http.StatusInternalServerError)
        return
    }
    defer f.Close()

    // Use the first sheet "Sheet1"
    sheetName := "Sheet1"
    rows, err := f.GetRows(sheetName)
    if err != nil {
        log.Println("Error reading Excel file rows:", err)
        http.Error(w, "Error reading Excel file", http.StatusInternalServerError)
        return
    }

    for i, row := range rows {
        if i == 0 {
            continue // Skip header row
        }

        // Prepare values, handling empty strings and missing columns, and trimming whitespace
        var m3Style, m3Sku, m3SkuDescription, m3ColourCode, m3ColourDescription string
        var m3SizeCode, m3SizeDescription, allesonSkuAlias, gtinAlias, upcAlias string
        var m3StdCasePackSize, cartonLength, cartonWidth, cartonHeight int
        var pieceWeight float64
        var mmrgdt string

        if len(row) > 0 {
            m3Style = strings.TrimSpace(row[0])
        }
        if len(row) > 1 {
            m3Sku = strings.TrimSpace(row[1])
        }
        if len(row) > 2 {
            m3SkuDescription = strings.TrimSpace(row[2])
        }
        if len(row) > 3 {
            m3ColourCode = strings.TrimSpace(row[3])
        }
        if len(row) > 4 {
            m3ColourDescription = strings.TrimSpace(row[4])
        }
        if len(row) > 5 {
            m3SizeCode = strings.TrimSpace(row[5])
        }
        if len(row) > 6 {
            m3SizeDescription = strings.TrimSpace(row[6])
        }
        if len(row) > 7 {
            allesonSkuAlias = strings.TrimSpace(row[7])
        }
        if len(row) > 8 {
            gtinAlias = strings.TrimSpace(row[8])
        }
        if len(row) > 9 {
            upcAlias = strings.TrimSpace(row[9])
        }
        if len(row) > 10 {
            m3StdCasePackSize, _ = strconv.Atoi(strings.TrimSpace(row[10]))
        }
        if len(row) > 11 {
            pieceWeight, _ = strconv.ParseFloat(strings.TrimSpace(row[11]), 64)
        }
        if len(row) > 12 {
            mmrgdt = strings.TrimSpace(row[12])
        }
        if len(row) > 13 {
            cartonLength, _ = strconv.Atoi(strings.TrimSpace(row[13]))
        }
        if len(row) > 14 {
            cartonWidth, _ = strconv.Atoi(strings.TrimSpace(row[14]))
        }
        if len(row) > 15 {
            cartonHeight, _ = strconv.Atoi(strings.TrimSpace(row[15]))
        }

        log.Printf("Inserting row %d: %v\n", i, row)

        // Start a new transaction for each row
        tx, err := db.GetDB().Begin()
        if err != nil {
            log.Println("Error beginning transaction:", err)
            http.Error(w, "Error beginning transaction", http.StatusInternalServerError)
            return
        }

        stmt, err := tx.Prepare(`
            INSERT INTO Badger (
                M3_Style, M3_Sku, M3_Sku_Description, M3_Colour_Code, M3_Colour_Description, 
                M3_Size_Code, M3_Size_Description, Alleson_Sku_Alias, GTIN_Alias, UPC_Alias, 
                M3_std_case_pack_size, Piece_Weight, MMRGDT, Carton_Length, Carton_Width, 
                Carton_Height
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
        `)
        if err != nil {
            tx.Rollback()
            log.Printf("Error preparing statement: %v", err)
            http.Error(w, "Error preparing statement", http.StatusInternalServerError)
            return
        }

        _, err = stmt.Exec(
            m3Style, m3Sku, m3SkuDescription, m3ColourCode, m3ColourDescription,
            m3SizeCode, m3SizeDescription, allesonSkuAlias, gtinAlias, upcAlias,
            m3StdCasePackSize, pieceWeight, mmrgdt, cartonLength, cartonWidth, cartonHeight,
        )
        if err != nil {
            if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
                log.Printf("Skipping duplicate GTIN_Alias: %s\n", gtinAlias)
                tx.Rollback()
                continue
            }
            tx.Rollback()
            log.Println("Error inserting data into database:", err)
            http.Error(w, "Error inserting data", http.StatusInternalServerError)
            return
        }

        err = tx.Commit()
        if err != nil {
            log.Println("Error committing transaction:", err)
            http.Error(w, "Error committing transaction", http.StatusInternalServerError)
            return
        }
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Data imported successfully"})
}

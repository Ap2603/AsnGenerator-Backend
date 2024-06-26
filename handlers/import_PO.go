package handlers

import (
	"AsnGenerator-Backend/db"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

func sendEmail(poNumber string, skippedLineNumber int, existingLineNumber int, emailAddresses []string) error {
    from := "apparelsfivestar@gmail.com"
    password := "abhd urol pvol cjlw"

    // Set up authentication information.
    auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")

    to := emailAddresses
    msg := []byte("To: " + strings.Join(emailAddresses, ", ") + "\r\n" +
        "Subject: Duplicate Entry Detected\r\n" +
        "\r\n" +
        "DURING IMPORT OF PO FILES, LINE NUMBER " + strconv.Itoa(skippedLineNumber) + 
        " in PO NUMBER " + poNumber + " was skipped. Quantities merged into LINE NUMBER " + 
        strconv.Itoa(existingLineNumber) + ".\r\n")

    err := smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)
    if err != nil {
        log.Printf("Error sending email to %v: %v", emailAddresses, err)
        return fmt.Errorf("error sending email: %w", err)
    }
    return nil
}

func ImportPOHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Parse the multipart form
    err := r.ParseMultipartForm(100 << 20) // 100 MB max memory
    if err != nil {
        log.Println("Error parsing form:", err)
        http.Error(w, "Error parsing form", http.StatusInternalServerError)
        return
    }

    emailAddresses := []string{"aamir.parekh36@gmail.com", "asif@fivestarapparels.com.pk"}

    files := r.MultipartForm.File["file"]
    if len(files) == 0 {
        http.Error(w, "No files uploaded", http.StatusBadRequest)
        return
    }

    for _, fileHeader := range files {
        file, err := fileHeader.Open()
        if err != nil {
            log.Println("Error retrieving file:", err)
            http.Error(w, "Error retrieving file", http.StatusInternalServerError)
            return
        }
        defer file.Close()

        // Save the file locally
        tempFile, err := os.Create(filepath.Join(os.TempDir(), fileHeader.Filename))
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

        for _, sheetName := range f.GetSheetMap() {
            if strings.ToLower(sheetName) == "summary" {
                continue // Skip summary sheets
            }

            rows, err := f.GetRows(sheetName)
            if err != nil {
                log.Println("Error reading sheet:", err)
                http.Error(w, "Error reading sheet", http.StatusInternalServerError)
                return
            }

            if len(rows) <= 7 {
                continue // Skip sheets with insufficient rows
            }

            poNumber := strings.TrimSpace(rows[4][0]) // Get the table name from cell A5
            total := strings.TrimSpace(rows[len(rows)-1][6]) // Get the total from the last row, cell G

            log.Printf("Inserting into PO: PO_Number: %s, Total: %s\n", poNumber, total)

            // Insert into PO table outside the transaction
            _, err = db.GetDB().Exec(`
                INSERT INTO PO (PO_Number, Total)
                VALUES ($1, $2)
                ON CONFLICT (PO_Number) DO NOTHING
            `, poNumber, total)
            if err != nil {
                log.Println("Error inserting into PO table:", err)
                http.Error(w, "Error inserting into PO table", http.StatusInternalServerError)
                return
            }

            var lastInsertedLine int
            duplicates := make([][]interface{}, 0)

            // Insert data into ItemsOrdered table
            for rowIndex, row := range rows[6 : len(rows)-1] { // Skip the first 6 rows and last row
                lineNumber, _ := strconv.Atoi(strings.TrimSpace(row[0]))
                itemNumber := strings.TrimSpace(row[1])
                style := strings.TrimSpace(row[2])
                colourSize := strings.TrimSpace(row[3])
                cost := strings.TrimSpace(row[4])
                pcs, pcsErr := strconv.Atoi(strings.TrimSpace(row[5]))
                if pcsErr != nil {
                    log.Printf("Error converting PCS to integer at row %d: %v\n", rowIndex+7, pcsErr)
                }
                total := strings.TrimSpace(row[6])
                exFacDate := strings.TrimSpace(row[7])

                log.Printf("Inserting into ItemsOrdered: PO_Number: %s, Line_Number: %d, Item_Number: %s, Style: %s, Colour_Size: %s, Cost: %s, Pcs: %d, Total: %s, Ex_Fac_Date: %s\n",
                    poNumber, lineNumber, itemNumber, style, colourSize, cost, pcs, total, exFacDate)

                // Begin a transaction for each insert
                tx, err := db.GetDB().Begin()
                if err != nil {
                    log.Println("Error beginning transaction:", err)
                    http.Error(w, "Error beginning transaction", http.StatusInternalServerError)
                    return
                }

                // Prepare statement for ItemsOrdered
                stmt, err := tx.Prepare(`
                    INSERT INTO ItemsOrdered (
                        PO_Number, Line_Number, Item_Number, Style, Colour_Size, Cost, Pcs, Total, Ex_Fac_Date
                    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
                `)
                if err != nil {
                    log.Println("Error preparing statement:", err)
                    tx.Rollback()
                    http.Error(w, "Error preparing statement", http.StatusInternalServerError)
                    return
                }

                _, err = stmt.Exec(poNumber, lineNumber, itemNumber, style, colourSize, cost, pcs, total, exFacDate)
                stmt.Close()
                if err != nil {
                    if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
                        // Add to duplicates list
                        duplicates = append(duplicates, []interface{}{pcs, total, poNumber, itemNumber, lineNumber, lastInsertedLine})
                        tx.Rollback()
                        continue
                    }
                    log.Printf("Error inserting into ItemsOrdered table at row %d: %v\n", rowIndex+7, row)
                    log.Println("Error inserting into ItemsOrdered table:", err)
                    tx.Rollback()
                    http.Error(w, "Error inserting into ItemsOrdered table", http.StatusInternalServerError)
                    return
                }

                lastInsertedLine = lineNumber

                err = tx.Commit()
                if err != nil {
                    log.Println("Error committing transaction:", err)
                    http.Error(w, "Error committing transaction", http.StatusInternalServerError)
                    return
                }
            }

            // Handle duplicates separately
            for _, duplicate := range duplicates {
                pcs, total, poNumber, itemNumber, lineNumber, lastInsertedLine := duplicate[0].(int), duplicate[1].(string), duplicate[2].(string), duplicate[3].(string), duplicate[4].(int), duplicate[5].(int)
                updateErr := updatePCS(pcs, total, poNumber, itemNumber, emailAddresses, lineNumber, lastInsertedLine)
                if updateErr != nil {
                    http.Error(w, "Error updating PCS", http.StatusInternalServerError)
                    return
                }
            }
        }
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Data imported successfully"})
}

func updatePCS(pcs int, total string, poNumber, itemNumber string, emailAddresses []string, lineNumber int, lastInsertedLine int) error {
    tx, err := db.GetDB().Begin()
    if err != nil {
        log.Println("Error beginning transaction for PCS update:", err)
        return err
    }
    defer tx.Rollback()

    // Remove any non-numeric characters from the total value
    re := regexp.MustCompile(`[^\d\.]`)
    cleanedTotal := re.ReplaceAllString(total, "")

    // Convert cleaned total to a float for addition
    newTotal, err := strconv.ParseFloat(cleanedTotal, 64)
    if err != nil {
        log.Println("Error converting total to float:", err)
        return err
    }

    // Get the existing total value
    row := tx.QueryRow(`
        SELECT Total
        FROM ItemsOrdered
        WHERE PO_Number = $1 AND Item_Number = $2
    `, poNumber, itemNumber)
    var existingTotal string
    err = row.Scan(&existingTotal)
    if err != nil {
        log.Println("Error retrieving existing total:", err)
        return err
    }

    existingTotalFloat, err := strconv.ParseFloat(re.ReplaceAllString(existingTotal, ""), 64)
    if err != nil {
        log.Println("Error converting existing total to float:", err)
        return err
    }

    // Calculate the new total
    combinedTotal := existingTotalFloat + newTotal

    // Format the combined total with a $ symbol
    formattedTotal := fmt.Sprintf("$%.2f", combinedTotal)

    _, err = tx.Exec(`
        UPDATE ItemsOrdered
        SET Pcs = Pcs + $1, Total = $2
        WHERE PO_Number = $3 AND Item_Number = $4
    `, pcs, formattedTotal, poNumber, itemNumber)
    if err != nil {
        log.Printf("Error updating PCS and Total for PO_Number: %s, Item_Number: %s, Add_Pcs: %d, New_Total: %s, Error: %v\n", poNumber, itemNumber, pcs, formattedTotal, err)
        return err
    }

    // Send email notification
    emailErr := sendEmail(poNumber, lineNumber, lastInsertedLine, emailAddresses)
    if emailErr != nil {
        log.Printf("Error sending email for PO_Number: %s, Item_Number: %s: %v\n", poNumber, itemNumber, emailErr)
        return emailErr
    }

    return tx.Commit()
}

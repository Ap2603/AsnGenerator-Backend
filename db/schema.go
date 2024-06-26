package db

import "log"

func CreateSchema() {
	schema := `  
    CREATE TABLE IF NOT EXISTS Badger (
        M3_Style VARCHAR(10),
        M3_Sku VARCHAR(30) UNIQUE,
        M3_Sku_Description VARCHAR(50),
        M3_Colour_Code VARCHAR(5),
        M3_Colour_Description VARCHAR(25),
        M3_Size_Code VARCHAR(25),
        M3_Size_Description VARCHAR(25),
        Alleson_Sku_Alias VARCHAR(30),
        GTIN_Alias VARCHAR(30) PRIMARY KEY,
        UPC_Alias VARCHAR(20),
        M3_std_case_pack_size INT,
        Piece_Weight DECIMAL,
        MMRGDT VARCHAR(20),
        Carton_Length INT,
        Carton_Width INT,
        Carton_Height INT
    );

    CREATE TABLE IF NOT EXISTS PO (
        PO_Number VARCHAR(20) PRIMARY KEY,
        Total VARCHAR(25)
    );


    CREATE TABLE IF NOT EXISTS ItemsOrdered (
        PO_Number VARCHAR(20),
        Line_Number INT,
        Item_Number VARCHAR(30),
        Style VARCHAR(50),
        Colour_Size VARCHAR(30),
        Cost VARCHAR(20),
        Pcs INT,
        Total VARCHAR(20),
        Ex_Fac_Date VARCHAR(25),
        PRIMARY KEY (PO_Number, Item_Number),
        FOREIGN KEY (PO_Number) REFERENCES PO(PO_Number),
        FOREIGN KEY (Item_Number) REFERENCES Badger(M3_Sku)
    );

    CREATE TABLE IF NOT EXISTS ShipmentID (
        ShipmentID VARCHAR(20) PRIMARY KEY
    );

    CREATE TABLE IF NOT EXISTS ASN (
        SSCC VARCHAR(20) PRIMARY KEY,
        Item_Code VARCHAR(30),
        Case_Pack_Size INT,
        PO_Number VARCHAR(20),
        Line_Number INT,
        ShipmentID VARCHAR(20),
        FOREIGN KEY (PO_Number) REFERENCES PO(PO_Number),
        FOREIGN KEY (Item_Code) REFERENCES Badger(M3_Sku),
        FOREIGN KEY (ShipmentID) REFERENCES ShipmentID(ShipmentID)
    );
    `
	_, err := DB.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}
}

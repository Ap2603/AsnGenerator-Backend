-- schema.sql

-- Create the Badger table
CREATE TABLE IF NOT EXISTS Badger (
    M3_Style VARCHAR(10),
    M3_Sku VARCHAR(30),
    M3_Sku_Description VARCHAR(50),
    M3_Colour_Code VARCHAR(5),
    M3_Colour_Description VARCHAR(10),
    M3_Size_Code VARCHAR(10),
    M3_Size_Description VARCHAR(10),
    Alleson_Sku_Alias VARCHAR(30),
    GTIN_Alias VARCHAR(30),
    UPC_Alias VARCHAR(20),
    M3_std_case_pack_size INT,
    Piece_Weight DECIMAL,
    MMRGDT VARCHAR(10),
    Carton_Length INT,
    Carton_Width INT,
    Carton_Height INT,
    PRIMARY KEY (GTIN_Alias)
);

-- Create the Asn table
CREATE TABLE IF NOT EXISTS Asn (
    item_code VARCHAR(15),
    SSCC VARCHAR(20),
    table_name VARCHAR(20),
    Line_number INTEGER,
    PRIMARY KEY (SSCC)
);

-- Create the UserAccounts table
CREATE TABLE IF NOT EXISTS UserAccounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE,
    password TEXT,
    is_admin BOOLEAN
);

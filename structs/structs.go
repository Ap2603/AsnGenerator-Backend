package structs

import "github.com/dgrijalva/jwt-go"

type BarcodeRequest struct {
	ShipmentID string `json:"shipment_id"`
	PONumber   string `json:"po_number"`
	GTIN       string `json:"gtin"`
	SSCC       string `json:"sscc"`
	Role       string `json:"role"`
	Override   bool   `json:"override"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

type AddShipmentIDRequest struct {
	ShipmentID string `json:"shipment_id"`
}


type POresults struct {
	LineNumber  int    `json:"line_number"`
	ItemNumber  string `json:"item_number"`
	Style       string `json:"style"`
	ColourSize  string `json:"colour_size"`
	Cost        string `json:"cost"`
	Pcs         int    `json:"pcs"`
	Total       string `json:"total"`
	ExFacDate   string `json:"ex_fac_date"`
}

type ASNResults struct {
	SSCC         string `json:"sscc"`
	ItemCode     string `json:"item_code"`
	CasePackSize int    `json:"case_pack_size"`
	PONumber     string `json:"po_number"`
	LineNumber   int    `json:"line_number"`
}

type EntryRemoval struct {
	SSCC      string `json:"sscc"`
	PONumber  string `json:"po_number"`
	ItemCode  string `json:"item_code"`
	LineNumber int    `json:"line_number"`
}

type POentries struct {
	PONumber   string `json:"poNumber"`
	LineNumber int    `json:"lineNumber"`
	ItemNumber string `json:"itemNumber"`
	Style      string `json:"style"`
	ColourSize string `json:"colourSize"`
	Cost       string `json:"cost"`
	Pcs        int    `json:"pcs"`
	Total      string `json:"total"`
	ExFacDate  string `json:"exFacDate"`
}
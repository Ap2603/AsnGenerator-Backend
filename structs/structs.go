package structs

import "github.com/dgrijalva/jwt-go"

type BarcodeRequest struct {
	ShipmentID string `json:"shipment_id"`
	PONumber   string `json:"po_number"`
	GTIN       string `json:"gtin"`
	SSCC       string `json:"sscc"`
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

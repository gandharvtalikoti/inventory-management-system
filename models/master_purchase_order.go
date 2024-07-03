package models

type MPO struct {
	MPOID         int    `json:"mpo_id"`
	PDFFilename   string `json:"pdf_filename"`
	InvoiceNumber string `json:"invoice_number"`
	EntityID      string `json:"entity_id"` 
}

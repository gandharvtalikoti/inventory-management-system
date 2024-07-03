package models

type MPO struct {
	MPOID           int    `json:"mpo_id"`
	PDFFilename     string `json:"pdf_filename"`
	InvoiceNumber   string `json:"invoice_number"`
	Mpo_instance_id string `json:"mpo_instance_id"`
}

type MPOInputParams struct {
	PDFFilename     string `json:"pdf_filename"`
	InvoiceNumber   string `json:"invoice_number"`
	Mpo_instance_id string `json:"mpo_instance_id"`
}

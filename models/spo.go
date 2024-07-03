package models

import (
	"inventory-management-system/models"
	"time"
)

type SPO struct {
	SPOID       int       `json:"spo_id"`
	MPOID       int       `json:"mpo_id"`
	InstanceID  string    `json:"instance_id"`
	WarehouseID string    `json:"warehouse_id"`
	DOA         time.Time `json:"doa"`
	Status      string    `json:"status"`
}

type SPOparams struct {
	// MPOID           int       `json:"mpo_id"`
	// InstanceID      string    `json:"instance_id"`
	// WarehouseID     string    `json:"warehouse_id"`
	// DOA             time.Time `json:"doa"`
	// Status          string    `json:"status"`
	// PDFFilename     string    `json:"pdf_filename"`
	// InvoiceNumber   string    `json:"invoice_number"`
	// Mpo_instance_id string    `json:"mpo_instance_id"`
	Mpo models.MPOparams                      `json:"mpo"`
	Spo models.SPOparams                      `json:"spo"`
	Sku []models.PurchaseOrderInventoryParams `json:"sku"`
}

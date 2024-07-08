package models

type Inventory struct {
	SKUID          int    `json:"sku_id"`
	Batch          string `json:"batch"`
	WarehouseID    string `json:"warehouse_id"`
	InStock        int    `json:"in_stock"`
	PendingReceipt int    `json:"pending_receipt"`
	InTransit      int    `json:"in_transit"`
	Received       int    `json:"received"`
	Quarantine     int    `json:"quarantine"`
	Committed      int    `json:"committed"`
	Reserved       int    `json:"reserved"`
	Available      int    `json:"available"`
	Damaged        int    `json:"damaged"`
}

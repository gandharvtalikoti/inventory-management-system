package models

import "time"

type Transactions struct {
	TID         int       `json:"tid"`
	SKUID       int       `json:"sku_id"`
	SPOID       int       `json:"spo_id"`
	SOID        int       `json:"so_id"`
	WarehouseID string    `json:"warehouse_id"`
	BinID       string    `json:"bin_id"`
	Qty         int       `json:"qty"`
	Batch       string    `json:"batch"`
	Type        string    `json:"type"`
	Source      string    `json:"source"`
	ExpiryDate  time.Time `json:"expiry_date"`
}

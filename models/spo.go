package models

import (
	"time"
)

type SPO struct {
	SPOID         int       `json:"spo_id"`
	MPOID         int       `json:"mpo_id"`
	SpoInstanceId string    `json:"spo_instance_id"`
	WarehouseID   string    `json:"warehouse_id"`
	DOA           time.Time `json:"doa"`
	Status        string    `json:"status"`
}

type SPOInputParams struct {
	SpoInstanceId string    `json:"spo_instance_id"`
	WarehouseID   string    `json:"warehouse_id"`
	DOA           time.Time `json:"doa"`
	Status        string    `json:"status"`
}

// this will be input for createSPO function
type SPOparams struct {
	Mpo          MPOInputParams                      `json:"mpo"`
	Spo          SPOInputParams                      `json:"spo"`
	Po_inventory []PurchaseOrderInventoryInputParams `json:"po_inventory"`
}

type AddNewSpoInputParams struct {
	MpiInstanceId string                              `json:"mpo_id"`
	Spo           SPOInputParams                      `json:"spo"`
	Po_inventory  []PurchaseOrderInventoryInputParams `json:"po_inventory"`
}

type UpdateSpoInputParams struct {
	SpoInstanceId string    `json:"spo_instance_id"`
	WarehouseID   string    `json:"warehouse_id"`
	BinID         string    `json:"bin_id"`
	DOA           time.Time `json:"doa"`
	Status        string    `json:"status"`
}

type CancleSpoInputParams struct {
	Spo SPOInputParams `json:"spo"`
}

type SKUToStock struct {
	SpoInstanceId string `json:"spo_instance_id"`
	SkuID         string `json:"sku_id"`
	Qty           int    `json:"qty"`
	WarehouseID   string `json:"warehouse_id"`
	BinID         string `json:"bin_id"`
}
type StockSpoInputParams struct {
	SpoInstanceId string       `json:"spo_instance_id"`
	StockSpoArray []SKUToStock `json:"sku_to_stock"`
}

type SplitSPO struct {
	SPO SPOInputParams                      `json: spo`
	SKU []PurchaseOrderInventoryInputParams `json:"sku"`
}
type SplitSPOInputParams struct {
	MPOInstanceID    string     `json:"mpo_instance_id"`
	OldSPOInstanceID string     `json:"old_spo_instance_id"`
	SplitSPO         []SplitSPO `json:"split_spo"`
}
package models

type PurchaseOrderInventory struct {
	POIID int    `json:"poi_id"`
	SKUID int    `json:"sku_id"`
	SPOID int    `json:"spo_id"`
	Qty   int    `json:"qty"`
	Batch string `json:"batch"`
}

type PurchaseOrderInventoryParams struct {
	Sku_instance_id string `json:"sku_instance_id"`
	Qty             int    `json:"qty"`
	Batch           string `json:"batch"`
}
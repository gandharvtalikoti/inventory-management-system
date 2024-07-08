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
	DOA           time.Time `json:"doa"`
	Status        string    `json:"status"`
}

type CancleSpoInputParams struct {
	Spo SPOInputParams `json:"spo"`
}

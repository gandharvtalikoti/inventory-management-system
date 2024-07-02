package models

import "time"

type PurchaseOrderInventory struct {
    POI_ID        string    `json:"poiId" gorm:"primaryKey"`
    PoID          string    `json:"poId"`
    SKUID         string    `json:"skuId"`
    DateOfArrival time.Time `json:"dateOfArrival"`
    Qty           int       `json:"qty"`
    Status        string    `json:"status"`
}
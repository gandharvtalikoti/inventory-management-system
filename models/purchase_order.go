package models

import "time"

type PurchaseOrder struct {
    PoID          string    `json:"poId" gorm:"primaryKey"`
    InvoiceNo     string    `json:"invoiceNo"`
    PoInstanceID  string    `json:"poInstanceId"`
    WareHouseID   string    `json:"warehouseId"`
    DateOfArrival time.Time `json:"dateOfArrival"`
    Status        string    `json:"status"`
}
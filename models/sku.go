package models

type SKU struct {
    SKUID       string `json:"skuId" gorm:"primaryKey"`
    SKUEntityID string `json:"skuEntityId"`
}
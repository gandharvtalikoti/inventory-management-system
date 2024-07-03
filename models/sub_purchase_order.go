package models

import "time"

type SPO struct {
    SPOID       int       `json:"spo_id"`
    MPOID       int       `json:"mpo_id"`
    InstanceID  string    `json:"instance_id"`
    WarehouseID string    `json:"warehouse_id"`
    DOA         time.Time `json:"doa"`
    Status      string    `json:"status"`
}

package main

import (
	//"encoding/json"
	"fmt"
	"inventory-management-system/config"
	"inventory-management-system/database"
	"inventory-management-system/models"
	"log"
	"time"
	//"time"
	//"time"
	//"time"
)

func createSKU(sku_instanse_id string) int64 {
	query := `INSERT INTO SKU (sku_instance_id) VALUES ($1)`
	res, err := database.DB.Exec(query, sku_instanse_id)
	if err != nil {
		log.Fatalf("Error creating SKU: %v", err)
	}
	skuID, err := res.LastInsertId()
	if err != nil {
		fmt.Println("Error getting SKU ID: ", err)
	}
	return skuID
}

func createMPO(mpo models.MPOInputParams) (int, error) {
	query := `
    INSERT INTO MPO (pdf_filename, invoice_number, mpo_instance_id)
    VALUES ($1, $2, $3)
    RETURNING mpo_id`

	//var mpoID int // ID of the newly created MPO
	var mpoID int // ID of the newly created MPO
	err := database.DB.QueryRow(query, mpo.PDFFilename, mpo.InvoiceNumber, mpo.Mpo_instance_id).Scan(&mpoID)

	if err != nil {
		return 0, fmt.Errorf("error creating MPO: %w", err)
	}
	return mpoID, nil

}

func GetSKUId(sku_instance_id string) int {
	var skuID int
	err := database.DB.QueryRow("SELECT sku_id FROM SKU WHERE sku_instance_id = $1", sku_instance_id).Scan(&skuID)
	
	if err != nil {
		_ = fmt.Errorf("error checking SKU existence: %w", err)
	}
	if skuID == 0 {
		// error message
		fmt.Printf("SKU with instance ID %s does not exist\n", sku_instance_id)
	}
	return skuID
}

func createSPO(spoParams models.SPOparams) (int, error) {
	// Check if mpo_id in models.SPO is present in MPO table
	fmt.Println("check")
	var mpoId int
	err := database.DB.QueryRow("SELECT mpo_id FROM MPO WHERE mpo_instance_id = $1", spoParams.Mpo.Mpo_instance_id).Scan(&mpoId)
	if err != nil {
		_ = fmt.Errorf("error checking MPO existence: %w", err)
	}
	fmt.Printf("MPO ID: %d\n", mpoId)
	if mpoId == 0 {
		// mpo_id does not exist in MPO table, create MPO first
		newMPO := models.MPOInputParams{
			PDFFilename:     spoParams.Mpo.PDFFilename,
			InvoiceNumber:   spoParams.Mpo.InvoiceNumber,
			Mpo_instance_id: spoParams.Mpo.Mpo_instance_id,
		}
		createdMPOID, err := createMPO(newMPO)
		if err != nil {
			fmt.Errorf("error creating MPO: %w", err)
		}
		mpoId = createdMPOID
		fmt.Printf("Created MPO with ID: %d\n", createdMPOID)

	}
	// if mpo_id exists while creating spo exists in mpo table then create SPO with the mpo_id and insert into SPO table
	query := `
			INSERT INTO SPO (mpo_id, instance_id, warehouse_id, doa, status)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING spo_id`

	var spoID int // ID of the newly created SPO
	err = database.DB.QueryRow(query, mpoId, spoParams.Spo.SpoInstanceId, spoParams.Spo.WarehouseID, spoParams.Spo.DOA, spoParams.Spo.Status).Scan(&spoID)
	if err != nil {
		fmt.Errorf("error creating SPO: %w", err)
	}
	//return spoID, nil

	// Insert into POI table
	for _, poi := range spoParams.Po_inventory {
		// get the sku_id from sku table using sku_instance_id, if not there print err
		skuID := GetSKUId(poi.Sku_instance_id)
		fmt.Println("SKU ID: ", skuID)
		// Insert into POI table
		query := `
			INSERT INTO PO_Inventory (sku_id,spo_id, qty, batch)
			VALUES ($1, $2, $3, $4)`
		_, err = database.DB.Exec(query, skuID, spoID, poi.Qty, poi.Batch)
		if err != nil {
			fmt.Errorf("error creating POI: %w", err)
		}
	}
	fmt.Print("inserted in poi table for spo_id: ", spoID)
	return spoID, nil

}

func getMPO(mpoID int) (models.MPO, error) {
	query := `
    SELECT mpo_id, pdf_filename, invoice_number, mpo_instance_id
    FROM MPO
    WHERE mpo_id = $1`

	var mpo models.MPO // MPO to store the retrieved data
	err := database.DB.QueryRow(query, mpoID).Scan(&mpo.MPOID, &mpo.PDFFilename, &mpo.InvoiceNumber, &mpo.Mpo_instance_id)
	if err != nil {
		return models.MPO{}, fmt.Errorf("error getting MPO: %w", err)
	}
	return mpo, nil
}

func deleteSPO(SpoInstanceId string) bool {
	query := `DELETE FROM SPO WHERE instance_id = $1`
	_, err := database.DB.Exec(query, SpoInstanceId)
	if err != nil {
		return false
	}
	return true

}

func main() {
	// Load the configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Connect to the database
	if err := database.ConnectDatabase(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer database.CloseDatabase()

	// Example usage of CRUD operations
	// newMPO := models.MPOInputParams{
	// 	PDFFilename:     "example.pdf",
	// 	InvoiceNumber:   "INV123456",
	// 	Mpo_instance_id: "blabla12345",
	// }

	// Create a new SPO
	newSPO := models.SPOparams{
		Mpo: models.MPOInputParams{
			PDFFilename:     "example.pdf",
			InvoiceNumber:   "INV123456",
			Mpo_instance_id: "hhhhhhhh",
		},
		Spo: models.SPOInputParams{
			SpoInstanceId: "aewdw",
			WarehouseID:   "W12345",
			DOA:           time.Now(),
			Status:        "Pending",
		},
		Po_inventory: []models.PurchaseOrderInventoryInputParams{
			{
				Sku_instance_id: "osaidhi237e1821e9jdo2",
				Qty:             20,
				Batch:           "B12345",
			},
			{
				Sku_instance_id: "eoifhe89rfy4hf834uf9",
				Qty:             60,
				Batch:           "B12345",
			},
			{
				Sku_instance_id: "psaiuiuygfhfgiuyi2",
				Qty:             68,
				Batch:           "saderfe",
			},
		},
	}

	// Create a new MPO
	//createMPO(newMPO)
	// if err != nil {
	// 	log.Fatalf("Error creating MPO: %v", err)
	// }
	// fmt.Printf("Created MPO with ID: %d\n", createdMPOID)

	// // Get the created MPO
	// retrievedMPO, err := getMPO(createdMPOID)
	// if err != nil {
	// 	log.Fatalf("Error retrieving MPO: %v", err)
	// }
	// jsonMPO, err := json.Marshal(retrievedMPO)
	// if err != nil {
	// 	log.Fatalf("Error marshaling MPO to JSON: %v", err)
	// }
	// fmt.Printf("Retrieved MPO: %s\n", jsonMPO)

	createSPO(newSPO)

	//createSKU("eoifhe89rfy4hf834uf9")
	//deleteSPO("I12345")

}

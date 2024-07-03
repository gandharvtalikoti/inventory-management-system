package main

import (
	"encoding/json"
	"fmt"
	"inventory-management-system/config"
	"inventory-management-system/database"
	"inventory-management-system/models"
	"log"
	"time"
)

func createMPO(mpo models.MPO) (int, error) {
	query := `
    INSERT INTO MPO (pdf_filename, invoice_number, mpo_instance_id)
    VALUES ($1, $2, $3)
    RETURNING mpo_id`

	var mpoID int // ID of the newly created MPO
	err := database.DB.QueryRow(query, mpo.PDFFilename, mpo.InvoiceNumber, mpo.Mpo_instance_id).Scan(&mpoID)
	if err != nil {
		return 0, fmt.Errorf("error creating MPO: %w", err)
	}
	return mpoID, nil
}

func createSPO(spo models.SPOparams) (int, error) {
	// Check if mpo_id in models.SPO is present in MPO table
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM MPO WHERE mpo_instance_id = $1", spo.Mpo.Mpo_instance_id).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error checking MPO existence: %w", err)
	}

	if count > 0 {
		// mpo_id exists in MPO table, insert into SPO
		query := `
			INSERT INTO SPO (mpo_id, instance_id, warehouse_id, doa, status)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING spo_id`

		var spoID int // ID of the newly created SPO
		err = database.DB.QueryRow(query, spo., spo.InstanceID, spo.WarehouseID, spo.DOA, spo.Status).Scan(&spoID)
		if err != nil {
			return 0, fmt.Errorf("error creating SPO: %w", err)
		}
		return spoID, nil
	} else {
		// mpo_id does not exist in MPO table, create MPO first
		newMPO := models.MPO{
			PDFFilename:     "example.pdf",
			InvoiceNumber:   "INV123456",
			Mpo_instance_id: "I12345",
		}

		createdMPOID, err := createMPO(newMPO)
		if err != nil {
			return 0, fmt.Errorf("error creating MPO: %w", err)
		}

		// Update the MPOID in the SPO model
		spo.MPOID = createdMPOID

		// Insert into SPO
		query := `
			INSERT INTO SPO (mpo_id, instance_id, warehouse_id, doa, status)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING spo_id`

		var spoID int // ID of the newly created SPO
		err = database.DB.QueryRow(query, spo.MPOID, spo.InstanceID, spo.WarehouseID, spo.DOA, spo.Status).Scan(&spoID)
		if err != nil {
			return 0, fmt.Errorf("error creating SPO: %w", err)
		}
		return spoID, nil
	}
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
	newMPO := models.MPO{
		PDFFilename:     "example.pdf",
		InvoiceNumber:   "INV123456",
		Mpo_instance_id: "I12345",
	}

	// Example usage of CRUD operations for SPO

	// Create a new SPO
newSPO := models.SPOparams{
		Mpo: models.MPOInputParams{
			PDFFilename:     "example.pdf",
			InvoiceNumber:   "INV123456",
			Mpo_instance_id: "I12345",
		},
		Spo: models.SPOInputParams{
			InstanceID:  "I12345",
			WarehouseID: "W12345",
			DOA:         time.Now(),
			Status:      "Pending",
		},
		Po_inventory: []models.PurchaseOrderInventoryInputParams{
			{
				Sku_instance_id: "osaidhi237e1821e9jdo2",
				Qty:  10,
				Batch: "B12345",
			},
			{
				Sku_instance_id: "eoifhe89rfy4hf834uf9-23",
				Qty:  10,
				Batch: "B12345",
			},
		},


	}
	createSPO(newSPO)
	// Create a new MPO
	createdMPOID, err := createMPO(newMPO)
	if err != nil {
		log.Fatalf("Error creating MPO: %v", err)
	}
	fmt.Printf("Created MPO with ID: %d\n", createdMPOID)

	// Get the created MPO
	retrievedMPO, err := getMPO(createdMPOID)
	if err != nil {
		log.Fatalf("Error retrieving MPO: %v", err)
	}
	jsonMPO, err := json.Marshal(retrievedMPO)
	if err != nil {
		log.Fatalf("Error marshaling MPO to JSON: %v", err)
	}
	fmt.Printf("Retrieved MPO: %s\n", jsonMPO)
	createSPO(newSPO)
}

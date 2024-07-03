package main

import (
	"encoding/json"
	"fmt"
	"inventory-management-system/config"
	"inventory-management-system/database"
	"inventory-management-system/models"
	"log"
)

func createMPO(mpo models.MPO) (int, error) {
    query := `
    INSERT INTO MPO (pdf_filename, invoice_number, entity_id)
    VALUES ($1, $2, $3)
    RETURNING mpo_id`

    var mpoID int // ID of the newly created MPO
    err := database.DB.QueryRow(query, mpo.PDFFilename, mpo.InvoiceNumber, mpo.EntityID).Scan(&mpoID)
    if err != nil {
        return 0, fmt.Errorf("error creating MPO: %w", err)
    }
    return mpoID, nil
}

func getMPO(mpoID int) (models.MPO, error) {
    query := `
    SELECT mpo_id, pdf_filename, invoice_number, entity_id
    FROM MPO
    WHERE mpo_id = $1`

    var mpo models.MPO // MPO to store the retrieved data
    err := database.DB.QueryRow(query, mpoID).Scan(&mpo.MPOID, &mpo.PDFFilename, &mpo.InvoiceNumber, &mpo.EntityID)
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
        PDFFilename:   "example.pdf",
        InvoiceNumber: "INV123456",
        EntityID:      "E12345",
    }

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
}
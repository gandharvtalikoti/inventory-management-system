package main

import (
	//"encoding/json"
	"database/sql"
	"fmt"
	"inventory-management-system/config"
	"inventory-management-system/database"
	"inventory-management-system/models"
	"time"

	//"time"

	"github.com/gofiber/fiber/v2/log"
)

func CreateSKU(sku_instanse_id string) (int64, error) {
	var existingID int64
	err := database.DB.QueryRow("SELECT sku_id FROM SKU WHERE sku_instance_id = $1", sku_instanse_id).Scan(&existingID)
	if err == nil {
		// sku_instance_id already exists, return the existing sku_id
		return existingID, nil
	} else if err != sql.ErrNoRows {
		// An error occurred during the query
		log.Fatalf("Error checking SKU existence: %v", err)
		return -1, err
	}

	// If sku_instance_id doesn't exist, proceed with insertion
	query := `INSERT INTO SKU (sku_instance_id) VALUES ($1) RETURNING sku_id`
	var skuID int64
	err = database.DB.QueryRow(query, sku_instanse_id).Scan(&skuID)
	if err != nil {
		log.Fatalf("Error creating SKU: %v", err)
		return -1, err
	}
	return skuID, nil
}
func DeleteSKU(sku_instance_id string) (bool, error) {
	var existingID int64
	err := database.DB.QueryRow("SELECT sku_id FROM SKU WHERE sku_instance_id = $1", sku_instance_id).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("SKU with sku_instance_id '%s' not found", sku_instance_id)
			return false, nil
		}
		log.Errorf("Error checking SKU existence: %v", err)
		return false, err
	}

	// If sku_instance_id exists, proceed with deletion
	query := `DELETE FROM SKU WHERE sku_instance_id = $1`
	_, err = database.DB.Exec(query, sku_instance_id)
	if err != nil {
		log.Fatalf("Error deleting SKU: %v", err)
		return false, err
	}

	return true, nil
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

func UpdateSPO(updateSPOParams models.UpdateSpoInputParams) error {
	//Get spo using the given instance_id and store the warehouseID and status in memory and then update the warehouseID and status to the new warehouse and status.
	//With the current SPO_ID get all the po_inventory rows.
	//Iterate through each po_inventory and with that sku_id, batch and current warehouse_id get the Inventory row. ---> For that particular row if the warehouse is new --> subtract that qty from the old row(check if qty matches) and then add that qty to the new row
	//else if warehouse is same --> subtract that qty(check if qty matches) and then update that same qty to the new column name

	//GET SPO using SPO Instance ID
	spoInstanceID := updateSPOParams.SpoInstanceId
	var spo models.SPO
	getspo_query := `
        SELECT spo_id, mpo_id, spo_instance_id, warehouse_id, doa, status
              FROM spo
              WHERE spo_instance_id = $1
        `
	spo_row := database.DB.QueryRow(getspo_query, spoInstanceID)

	row_err := spo_row.Scan(&spo.SPOID, &spo.MPOID, &spo.SpoInstanceId, &spo.WarehouseID, &spo.DOA, &spo.Status)
	if row_err != nil {
		if row_err == sql.ErrNoRows {
			log.Errorf("No row found for the given SPO instance id: ", row_err)
			return row_err
		}
		return row_err
	}

	currentWarehouse := spo.WarehouseID
	CurrentStatus := spo.Status

	var newWarehouse bool
	if updateSPOParams.WarehouseID != currentWarehouse {
		newWarehouse = true
	}

	//Get Purchase Order Inventory Rows
	get_poInventory_query := `
        SELECT poi_id, sku_id, spo_id, qty, batch
              FROM po_inventory
              WHERE spo_id = $1
    `

	log.Info("SPO ID: ", spo.SPOID)
	poInventoryRows, err := database.DB.Query(get_poInventory_query, spo.SPOID)
	if err != nil {
		log.Errorf("Error Getting Purchase Order Inventory: ", err)
		return err
	}

	defer poInventoryRows.Close()

	var pois []models.PurchaseOrderInventory
	for poInventoryRows.Next() {
		var poi models.PurchaseOrderInventory
		if err := poInventoryRows.Scan(&poi.POIID, &poi.SKUID, &poi.SPOID, &poi.Qty, &poi.Batch); err != nil {
			log.Errorf("Error Getting Purchase Order Inventory Row: ", err)
			return err
		}
		pois = append(pois, poi)
	}

	//Get and Update Inventory for each SKU's
	get_inventory_query := `
    SELECT inv_id, sku_id, batch, warehouse_id, bin_id, in_stock, pending_receipt, in_transit, received, quarantine, committed, reserved, available, damaged
              FROM inventory
              WHERE sku_id = $1 AND batch = $2 AND warehouse_id = $3
    `
	updateSPO := false
	for _, poi := range pois {
		inventory_row := database.DB.QueryRow(get_inventory_query, poi.SKUID, poi.Batch, currentWarehouse)
		var inventory models.Inventory
		var invID int64
		inventory_row_err := inventory_row.Scan(&invID, &inventory.SKUID, &inventory.Batch, &inventory.WarehouseID,
			&inventory.InStock, &inventory.PendingReceipt, &inventory.InTransit, &inventory.Received, &inventory.Quarantine,
			&inventory.Committed, &inventory.Reserved, &inventory.Available, &inventory.Damaged)
		if inventory_row_err != nil {
			if inventory_row_err == sql.ErrNoRows {
				log.Errorf("No Rows Found in Inventory: ", inventory_row_err)
				return inventory_row_err
			}
		}

		var currentValue int64
		get_currentValue_query := fmt.Sprintf(`
        SELECT %s
        FROM inventory
        WHERE sku_id = $1 AND batch = $2 AND warehouse_id = $3
    `, CurrentStatus)

		err := database.DB.QueryRow(get_currentValue_query, poi.SKUID, poi.Batch, currentWarehouse).Scan(&currentValue)
		if err != nil {
			log.Errorf("Error Getting Row")
			return err
		}

		if currentValue < int64(poi.Qty) {
			log.Errorf("Subtraction Not possible in Inventory")
			return err
		}
		if newWarehouse {

			//Subtract the QTY From the Old row entry
			updateQuery := fmt.Sprintf("UPDATE inventory SET %s = %s - $4 WHERE sku_id = $1 AND batch = $2 AND warehouse_id = $3", CurrentStatus, CurrentStatus)
			_, err := database.DB.Exec(updateQuery, poi.SKUID, poi.Batch, currentWarehouse, poi.Qty)
			if err != nil {
				log.Errorf("Error Deleting QTY in inventory")
				return err
			}

			//Add the value to the new row entry
			insertQuery := fmt.Sprintf(`
                INSERT INTO inventory (sku_id, warehouse_id, batch, %s) 
                VALUES ($1, $2, $3, $4)
            `, updateSPOParams.Status)

			_, insert_err := database.DB.Exec(insertQuery, poi.SKUID, updateSPOParams.WarehouseID, poi.Batch, poi.Qty)
			if insert_err != nil {
				log.Errorf("Error Inserting into inventory DB")
				return insert_err
			}

		} else {
			update_inventory_query := fmt.Sprintf(`UPDATE inventory
        SET 
            %s = COALESCE(%s, 0) + $4,
        %s = COALESCE(%s, 0) - $5
        WHERE sku_id = $1 AND batch = $2 AND warehouse_id = $3;`, updateSPOParams.Status, updateSPOParams.Status, CurrentStatus, CurrentStatus)

			_, update_inventory_err := database.DB.Exec(update_inventory_query, poi.SKUID, poi.Batch, currentWarehouse, poi.Qty, poi.Qty)
			if update_inventory_err != nil {
				log.Errorf("Error Updating Inventory: ", update_inventory_err)
				return update_inventory_err
			}

		}

		//Change the status of the po to current status
		if !updateSPO {
			update_spo_query := `
            UPDATE spo
    SET warehouse_id = $1, status = $2, doa = $3
    WHERE spo_instance_id = $4
        `
			_, update_spo_err := database.DB.Exec(update_spo_query, updateSPOParams.WarehouseID, updateSPOParams.Status, updateSPOParams.DOA, spoInstanceID)

			if update_spo_err != nil {
				log.Errorf("Error updating spo table: %v", err)
				return err
			}
			updateSPO = true
		}

	}
	return nil
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
			INSERT INTO SPO (mpo_id, spo_instance_id, warehouse_id, doa, status)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING spo_id`

	var spoID int // ID of the newly created SPO
	var spoStatus = spoParams.Spo.Status
	var spo_warehouse_id = spoParams.Spo.WarehouseID
	err = database.DB.QueryRow(query, mpoId, spoParams.Spo.SpoInstanceId, spo_warehouse_id, spoParams.Spo.DOA, spoStatus).Scan(&spoID)
	if err != nil {
		fmt.Errorf("error creating SPO: %w", err)
	}

	// Insert into POI table
	for _, poi := range spoParams.Po_inventory {
		// get the sku_id from sku table using sku_instance_id, if not there print err
		skuID := GetSKUId(poi.Sku_instance_id)
		fmt.Println("SKU ID: ", skuID)
		// Insert into POI table
		query := `
			INSERT INTO po_inventory (sku_id,spo_id, qty, batch)
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

func deleteSku(sku_instance_id string) bool {
	query := `DELETE FROM SKU WHERE sku_instance_id = $1`
	_, err := database.DB.Exec(query, sku_instance_id)
	if err != nil {
		return false
	}
	fmt.Println("Deleted SKU with instance ID: ", sku_instance_id)
	return true
}

func addSpo(addSpo models.AddNewSpoInputParams) (int, error) {
	// Check if mpo_id in models.SPO is present in MPO table
	fmt.Println("check addSpo")
	var mpoId int
	err := database.DB.QueryRow("SELECT mpo_id FROM MPO WHERE mpo_instance_id = $1", addSpo.MpiInstanceId).Scan(&mpoId)
	if err != nil {
		_ = fmt.Errorf("error checking MPO existence: %w", err)
	}
	fmt.Printf("MPO ID: %d\n", mpoId)
	if mpoId == 0 {
		fmt.Println("MPO with instance ID does not exist", addSpo.MpiInstanceId)
		return 0, nil
	}
	// if mpo_id exists while creating spo exists in mpo table then create SPO with the mpo_id and insert into SPO table
	query := `
			INSERT INTO SPO (mpo_id, instance_id, warehouse_id, doa, status)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING spo_id`

	var spoID int // ID of the newly created SPO
	err = database.DB.QueryRow(query, mpoId, addSpo.Spo.SpoInstanceId, addSpo.Spo.WarehouseID, addSpo.Spo.DOA, addSpo.Spo.Status).Scan(&spoID)
	if err != nil {
		fmt.Errorf("error creating SPO: %w", err)
	}
	//return spoID, nil

	// Insert into POI table
	for _, poi := range addSpo.Po_inventory {
		// get the sku_id from sku table using sku_instance_id, if not there print err
		skuID := GetSKUId(poi.Sku_instance_id)
		fmt.Println("SKU ID: ", skuID)
		// Insert into POI table
		query := ` INSERT INTO PO_Inventory (sku_id,spo_id, qty, batch)
			VALUES ($1, $2, $3, $4)`
		_, err = database.DB.Exec(query, skuID, spoID, poi.Qty, poi.Batch)
		if err != nil {
			fmt.Errorf("error creating POI: %w", err)
		}
	}
	return spoID, nil
}

func cancelSpo(cancelSpoData models.CancleSpoInputParams) error {
	// get spo from spo table using spo_instance_id
	var spoID int
	err := database.DB.QueryRow("SELECT spo_id FROM SPO WHERE instance_id = $1", cancelSpoData.Spo.SpoInstanceId).Scan(&spoID)
	if err != nil {
		_ = fmt.Errorf("error checking SPO existence: %w", err)
	}
	// using spoId get qty from po_inventory table and store it in memory
	var qty int
	err = database.DB.QueryRow("SELECT qty FROM PO_Inventory WHERE spo_id = $1", spoID).Scan(&qty)
	if err != nil {
		_ = fmt.Errorf("error checking PO_Inventory existence: %w", err)
	}
	if spoID == 0 {
		// error message
		fmt.Printf("SPO with instance ID %s does not exist\n", cancelSpoData.Spo.SpoInstanceId)
		return nil
	}

	// update the qty in inventory table with the column name same as status
	// get the status of the spo from spo table
	var status string
	err = database.DB.QueryRow("SELECT status FROM SPO WHERE spo_id = $1", spoID).Scan(&status)
	if err != nil {

		_ = fmt.Errorf("error checking SPO existence: %w", err)
	}

	// update the qty in inventory table with the column name same as status
	// subtract the qty from the inventory table with the column name same as status
	query := fmt.Sprintf("UPDATE inventory SET %s = %s - $1 WHERE sku_id = $2", status, status)
	_, err = database.DB.Exec(query, qty)
	if err != nil {
		return fmt.Errorf("error updating inventory: %w", err)
	}
	
	// if spo exist and update the status to cancelled if not return error
	u_query := `UPDATE SPO SET status = 'cancelled' WHERE AND spo_id = $1`
	_, err = database.DB.Exec(u_query, spoID)
	if err != nil {
		return fmt.Errorf("error updating SPO: %w", err)
	}
	return nil
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

	// newMPO := models.MPOInputParams{
	// 	PDFFilename:     "carpet456.pdf",
	// 	InvoiceNumber:   "INV-456",
	// 	Mpo_instance_id: "CARPET-456",
	// }
	// // Create a new MPO
	// createMPO(newMPO)



	// Create a SPO
	// newSPO := models.SPOparams{
	// 	Mpo: models.MPOInputParams{
	// 		PDFFilename:     "carpet456.pdf",
	// 		InvoiceNumber:   "INV-456",
	// 		Mpo_instance_id: "C1",
	// 	},
	// 	Spo: models.SPOInputParams{
	// 		SpoInstanceId: "SPO-3",
	// 		WarehouseID:   "W1",
	// 		DOA:           time.Now(),
	// 		Status:        "pending_reciept",
	// 	},
	// 	Po_inventory: []models.PurchaseOrderInventoryInputParams{
	// 		{
	// 			Sku_instance_id: "SKU-4",
	// 			Qty:             20,
	// 			Batch:           "BA1",
	// 		},
	// 		{
	// 			Sku_instance_id: "SKU-5",
	// 			Qty:             60,
	// 			Batch:           "BA2",
	// 		},
	// 		{
	// 			Sku_instance_id: "SKU-6",
	// 			Qty:             68,
	// 			Batch:           "BA3",
	// 		},
	// 	},
	// }
	// createSPO(newSPO)

	// addNewSpoToExistingMpo := models.AddNewSpoInputParams{
	// 	MpiInstanceId: "CARPET-456",
	// 	Spo: models.SPOInputParams{
	// 		SpoInstanceId: "SPo",
	// 		WarehouseID:   "W12345",
	// 		DOA:           time.Now(),
	// 		Status:        "Next status",
	// 	},
	// 	Po_inventory: []models.PurchaseOrderInventoryInputParams{
	// 		{
	// 			Sku_instance_id: "osaidhi237e1821e9jdo2",
	// 			Qty:             20,
	// 			Batch:           "B12345",
	// 		},
	// 		{
	// 			Sku_instance_id: "eoifhe89rfy4hf834uf9",
	// 			Qty:             60,
	// 			Batch:           "B12345",
	// 		},
	// 		{
	// 			Sku_instance_id: "psaiuiuygfhfgiuyi2",
	// 			Qty:             68,
	// 			Batch:           "saderfe",
	// 		},
	// 	},
	// }
	// addSpo(addNewSpoToExistingMpo)

	
	// cancel spo
	cancelSpoData := models.CancleSpoInputParams{
		Spo: models.SPOInputParams{
			SpoInstanceId: "SPO-3",
			WarehouseID:   "W1",
			DOA:           time.Now(),
			Status:        "pending_reciept",
		}}
	cancelSpo(cancelSpoData)

	// CreateSKU("SKU-11")
	
	

}

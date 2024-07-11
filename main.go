package main

import (
	//"encoding/json"
	"database/sql"
	"fmt"
	"inventory-management-system/config"
	"inventory-management-system/database"
	"inventory-management-system/models"

	"time"

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

func CreateMPO(mpo models.MPOInputParams) (int, error) {
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
	//Get spo using the given spo_instance_id and store the warehouseID and status in memory and then update the warehouseID and status to the new warehouse and status.
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
    SELECT inv_id, sku_id, batch, warehouse_id, bin_id, in_stock, pending_reciept, in_transit, received, quarantine, committed, reserved, available, damaged
              FROM inventory
              WHERE sku_id = $1 AND batch = $2 AND warehouse_id = $3
    `
	updateSPO := false
	for _, poi := range pois {
		inventory_row := database.DB.QueryRow(get_inventory_query, poi.SKUID, poi.Batch, currentWarehouse)
		var inventory models.Inventory
		var invID int64
		inventory_row_err := inventory_row.Scan(&invID, &inventory.SKUID, &inventory.Batch, &inventory.WarehouseID, &inventory.BinID,
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
			fmt.Printf("Error Getting Row: %v", err)
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
				log.Errorf("Error Updating Inventory:")
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

		} else { //If the warehouse is same then update the qty in the same row
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

		if updateSPOParams.Status == "in_stock" {
			// insert into transactions table for each sku
			transactionQuery := `
			INSERT INTO transactions (sku_id, spo_id, warehouse_id, qty ,batch, type, source, expiry_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`
			_, err = database.DB.Exec(transactionQuery, poi.SKUID, poi.SPOID, updateSPOParams.WarehouseID, poi.Qty, poi.Batch, "plus", "source", time.Now())
			if err != nil {
				log.Errorf("Error inserting into transactions table: %v", err)
				return err
			}
			fmt.Print("Inserted into transactions table\n")
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
		}

	} // end for loop
	return nil
}

// The `createSPO` function creates a new SPO (Supplier Purchase Order) along with associated MPO
// (Manufacturer Purchase Order) and POI (Purchase Order Inventory) entries in the database.
func createSPO(spoParams models.SPOparams) (int, error) {
	// Check if mpo_id in models.SPO is present in MPO table
	fmt.Println("create mpo check")
	var mpoId int
	err := database.DB.QueryRow("SELECT mpo_id FROM MPO WHERE mpo_instance_id = $1", spoParams.Mpo.Mpo_instance_id).Scan(&mpoId)
	if err != nil {
		_ = fmt.Errorf("error checking MPO existence: %w", err)
	}
	fmt.Printf("MPO ID(if 0 means no mpo): %d\n", mpoId)
	if mpoId == 0 {
		// mpo_id does not exist in MPO table, create MPO first
		newMPO := models.MPOInputParams{
			PDFFilename:     spoParams.Mpo.PDFFilename,
			InvoiceNumber:   spoParams.Mpo.InvoiceNumber,
			Mpo_instance_id: spoParams.Mpo.Mpo_instance_id,
		}
		createdMPOID, err := CreateMPO(newMPO)
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
			return -1, err
		}

		// insert into inventory table
		// check if sku_id and batch already exists in inventory table
		var invID int
		existing_inv_query := `SELECT inv_id FROM inventory WHERE sku_id = $1 AND warehouse_id = $2 AND batch = $3`
		existing_inv_err := database.DB.QueryRow(existing_inv_query, skuID, spo_warehouse_id, poi.Batch).Scan(&invID)
		if existing_inv_err != nil {
			if existing_inv_err == sql.ErrNoRows {
				// insert into inventory table
				insert_inv_query := `INSERT INTO inventory (sku_id, warehouse_id, batch, pending_reciept) VALUES ($1, $2, $3, $4)`
				_, insert_inv_err := database.DB.Exec(insert_inv_query, skuID, spoParams.Spo.WarehouseID, poi.Batch, poi.Qty)
				if insert_inv_err != nil {
					fmt.Errorf("error creating inventory: %w", insert_inv_err)
					return -1, insert_inv_err
				}
			} else {
				fmt.Errorf("error checking inventory existence: %w", existing_inv_err)
				return -1, existing_inv_err
			}
		} else {
			// update the inventory table
			updateQuery := `
        UPDATE inventory
        SET pending_reciept = pending_reciept + $1
        WHERE inv_id = $2`
			_, err = database.DB.Exec(updateQuery, poi.Qty, invID)
			if err != nil {
				log.Fatalf("Error updating inventory: %v", err)
				return -1, err
			}
		}
	}
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

// The function `deleteSPO` deletes a record from the SPO table based on the provided SpoInstanceId.
func deleteSPO(SpoInstanceId string) bool {
	query := `DELETE FROM SPO WHERE instance_id = $1`
	_, err := database.DB.Exec(query, SpoInstanceId)
	if err != nil {
		return false
	}
	return true
}

// The function `deleteSku` deletes a SKU from the database based on the provided SKU instance ID.
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
		//Insert into Inventory
		var invID int64
		//Check if an entry already exists with the same sku_id, warehouse_id, batch
		existing_inv_query := `
        SELECT inv_id
         FROM inventory
         WHERE sku_id = $1 AND warehouse_id = $2 AND batch = $3
        `

		existing_err := database.DB.QueryRow(existing_inv_query, skuID, addSpo.Spo.WarehouseID, poi.Batch).Scan(&invID)
		if existing_err != nil {
			if existing_err == sql.ErrNoRows {
				insertQuery := fmt.Sprintf(`
                INSERT INTO inventory (sku_id, warehouse_id, batch, %s)
                VALUES ($1, $2, $3, $4)
                RETURNING inv_id
                `, addSpo.Spo.Status)

				insert_query_err := database.DB.QueryRow(insertQuery, skuID, addSpo.Spo.WarehouseID, poi.Batch, poi.Qty).Scan(&invID)
				if insert_query_err != nil {
					log.Errorf("Error inserting Inventory: ", err)
					return -1, insert_query_err
				}
				log.Info("Newly added Inventory ID: ", invID)

			} else {
				log.Fatalf("Error checking inventory: %v", err)
				return -1, err
			}
		} else {

			//Row Exists, update the pending_receipt + $1
			updateQuery := fmt.Sprintf(`
        UPDATE inventory
        SET %s = COALESCE(%s, 0) + $1,
        WHERE inv_id = $2`, addSpo.Spo.Status, addSpo.Spo.Status)
			_, err = database.DB.Exec(updateQuery, poi.Qty, invID)
			if err != nil {
				log.Fatalf("Error updating inventory: %v", err)
				return -1, err
			}
		}
	}
	return spoID, nil
}

func CancelSPO(cancelSpoData models.CancleSpoInputParams) error {
	// get spo from spo table using spo_instance_id
	var spoID int
	var currentStatus string
	var warehouseID string
	err := database.DB.QueryRow("SELECT spo_id, status, warehouse_id FROM SPO WHERE spo_instance_id = $1", cancelSpoData.SpoInstanceId).Scan(&spoID, &currentStatus, &warehouseID)
	if err != nil {
		_ = fmt.Errorf("error checking SPO existence: %w", err)
	}
	if spoID == 0 {
		// error message
		fmt.Printf("SPO with instance ID %s does not exist\n", cancelSpoData.SpoInstanceId)
		return nil
	}

	var pois []models.PurchaseOrderInventory
	get_poInventory_query := `
		SELECT poi_id, sku_id, spo_id, qty, batch
			  FROM po_inventory
			  WHERE spo_id = $1
	`
	poInventoryRows, err := database.DB.Query(get_poInventory_query, spoID)
	if err != nil {
		log.Errorf("Error Getting Purchase Order Inventory: ", err)
		return err
	}
	defer poInventoryRows.Close()

	// get the po_inventory rows
	for poInventoryRows.Next() {
		var poi models.PurchaseOrderInventory
		if err := poInventoryRows.Scan(&poi.POIID, &poi.SKUID, &poi.SPOID, &poi.Qty, &poi.Batch); err != nil {
			log.Errorf("Error Getting Purchase Order Inventory Row: ", err)
			return err
		}
		pois = append(pois, poi)
	}

	// print the po_inventory rows
	fmt.Println("PO Inventory Rows: ", pois)

	// for every po_inventory row, update the qty in inventory table
	for _, poi := range pois {
		// update the qty in inventory table with the column name same as status

		// get the current qty from the inventory table
		var currentQty int
		get_currentQty_query := fmt.Sprintf("SELECT %s FROM inventory WHERE sku_id = $1 AND warehouse_id = $2 AND batch = $3", currentStatus)
		err := database.DB.QueryRow(get_currentQty_query, poi.SKUID, warehouseID, poi.Batch).Scan(&currentQty)
		if err != nil {
			return fmt.Errorf("error getting current qty from inventory: %w", err)
		}

		// subtract the qty from the inventory table with the column name same as status
		if currentQty < poi.Qty {
			return fmt.Errorf("error subtracting qty from inventory: %w", err)
		}

		query := fmt.Sprintf("UPDATE inventory SET %s = %s - $1 WHERE sku_id = $2 AND warehouse_id = $3 AND batch = $4", currentStatus, currentStatus)
		_, err = database.DB.Exec(query, poi.Qty, poi.SKUID, warehouseID, poi.Batch)
		if err != nil {
			return fmt.Errorf("error updating inventory: %w", err)
		}
	}

	// update the status to cancelled in spo table if not return error
	u_query := `UPDATE SPO SET status = 'cancelled' WHERE spo_id = $1`
	_, err = database.DB.Exec(u_query, spoID)
	if err != nil {
		return fmt.Errorf("error updating SPO: %w", err)
	}
	return nil
}

func StockSpo(stockSpoData models.StockSpoInputParams) error {

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

	//Create a SPO
	// newSPO := models.SPOparams{
	// 	Mpo: models.MPOInputParams{
	// 		PDFFilename:     "invoice.pdf",
	// 		InvoiceNumber:   "inv-123",
	// 		Mpo_instance_id: "C1",
	// 	},
	// 	Spo: models.SPOInputParams{
	// 		SpoInstanceId: "SPO-2",
	// 		WarehouseID:   "W1",
	// 		DOA:           time.Now(),
	// 		Status:        "pending_reciept",
	// 	},
	// 	Po_inventory: []models.PurchaseOrderInventoryInputParams{
	// 		{
	// 			Sku_instance_id: "SKU-1",
	// 			Qty:             111,
	// 			Batch:           "BA88",
	// 		},
	// 		{
	// 			Sku_instance_id: "SKU-3",
	// 			Qty:             222,
	// 			Batch:           "BA99",
	// 		},
	// 		{
	// 			Sku_instance_id: "SKU-4",
	// 			Qty:             333,
	// 			Batch:           "BA10",
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
	// cancelSpoData := models.CancleSpoInputParams{
	// 	SpoInstanceId: "SPO-1",
	// 	WarehouseID:   "W1",
	// 	DOA:           time.Now(),
	// 	Status:        "cancelled",
	// }

	// CancelSPO(cancelSpoData)

	// update spo
	// updateSPOParams := models.UpdateSpoInputParams{
	// 	SpoInstanceId: "SPO-1",
	// 	WarehouseID:   "W2",
	// 	DOA:           time.Now(),
	// 	Status:        "pending_reciept",
	// }

	// UpdateSPO(updateSPOParams)

	// stock spo
	stockSpoData := models.StockSpoInputParams{
		StockSpoArray: []models.SKUToStock{
			{
				
				Qty:         10,
				WarehouseID: "W1",
				BinID:       "B1",
			},
			{
				Qty:         20,
				WarehouseID: "W2",
				BinID:       "B2",
			},
		}}
	StockSpo(stockSpoData)

}

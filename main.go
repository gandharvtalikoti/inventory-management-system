package main

import (
	//"encoding/json"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"inventory-management-system/config"
	"inventory-management-system/database"
	"inventory-management-system/models"

	"time"

	"github.com/gofiber/fiber/v2/log"
)

func getSpoID(spo_instance_id string) int {
	query := `SELECT spo_id FROM SPO WHERE spo_instance_id = $1`
	var spoID int
	err := database.DB.QueryRow(query, spo_instance_id).Scan(&spoID)
	if err != nil {
		_ = fmt.Errorf("error checking SPO existence: %w", err)
	}
	if spoID == 0 {
		// error message
		fmt.Printf("SPO with instance ID %s does not exist\n", spo_instance_id)
	}
	return spoID
}

func getSPO(spo_instance_id string) (models.SPO, error) {
	query := `
	SELECT spo_id, mpo_id, spo_instance_id, warehouse_id, doa, status
	FROM SPO
	WHERE spo_instance_id = $1`

	var spo models.SPO // SPO to store the retrieved data
	err := database.DB.QueryRow(query, spo_instance_id).Scan(&spo.SPOID, &spo.MPOID, &spo.SpoInstanceId, &spo.WarehouseID, &spo.DOA, &spo.Status)
	if err != nil {
		return models.SPO{}, fmt.Errorf("error getting SPO: %w", err)
	}
	return spo, nil
}

func getPoInventories(spoID int) ([]models.PurchaseOrderInventory, error) {
	query := `
	SELECT poi_id, sku_id, spo_id, qty, batch
	FROM po_inventory
	WHERE spo_id = $1`

	poInventoryRows, err := database.DB.Query(query, spoID)
	if err != nil {
		log.Errorf("Error Getting Purchase Order Inventory: ", err)
		return nil, err
	}

	defer poInventoryRows.Close()

	var pois []models.PurchaseOrderInventory
	for poInventoryRows.Next() {
		var poi models.PurchaseOrderInventory
		if err := poInventoryRows.Scan(&poi.POIID, &poi.SKUID, &poi.SPOID, &poi.Qty, &poi.Batch); err != nil {
			log.Errorf("Error Getting Purchase Order Inventory Row: ", err)
			return nil, err
		}
		pois = append(pois, poi)
	}
	return pois, nil
}

func getPoi(SKUId int, spoID int) (models.PurchaseOrderInventory, error) {
	query := `
	SELECT poi_id, sku_id, spo_id, qty, batch
	FROM po_inventory
	WHERE sku_id = $1 AND spo_id = $2`

	var poi models.PurchaseOrderInventory
	err := database.DB.QueryRow(query, SKUId, spoID).Scan(&poi.POIID, &poi.SKUID, &poi.SPOID, &poi.Qty, &poi.Batch)
	if err != nil {
		return models.PurchaseOrderInventory{}, fmt.Errorf("error getting POI: %w", err)
	}
	return poi, nil

}
func getSKUId(sku_instance_id string) (int, error) {
	var skuID int
	err := database.DB.QueryRow("SELECT sku_id FROM SKU WHERE sku_instance_id = $1", sku_instance_id).Scan(&skuID)

	if err != nil {
		_ = fmt.Errorf("error checking SKU existence: %w", err)
		return -1, err
	}
	if skuID == 0 {
		// error message
		fmt.Printf("SKU with instance ID %s does not exist\n", sku_instance_id)
		return -1, err

	}
	return skuID, nil
}

func getMPO(mpoID int) (models.MPO, error) {
	query := `
    SELECT mpo_id, pdf_filename, invoice_number, mpo_instance_id
    FROM MPO
    WHERE mpo_id = $1`

	var mpo models.MPO // MPO to store the retrieved data
	err := database.DB.QueryRow(query, mpoID).Scan(&mpo.MPOID, &mpo.PDFFilename, &mpo.InvoiceNumber, &mpo.MPOInstanceID)
	if err != nil {
		return models.MPO{}, fmt.Errorf("error getting MPO: %w", err)
	}
	return mpo, nil
}

func CreateSKU(sku_instanse_id string, commit bool) (int64, error) {

	tx, err := database.DB.Begin()
	if err != nil {
		log.Errorf("Error starting transaction")
		return -1, err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			if commit {
				err = tx.Commit()
			} else {
				tx.Rollback()
			}
		}

	}()

	var existingID int64
	err = tx.QueryRow("SELECT sku_id FROM SKU WHERE sku_instance_id = $1", sku_instanse_id).Scan(&existingID)
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
	err = tx.QueryRow(query, sku_instanse_id).Scan(&skuID)
	if err != nil {
		log.Fatalf("Error creating SKU: %v", err)
		return -1, err
	}
	return skuID, nil
}

// func CreateSKU(sku_instanse_id string) (int64, error) {

// 	var existingID int64
// 	err := tx.QueryRow("SELECT sku_id FROM SKU WHERE sku_instance_id = $1", sku_instanse_id).Scan(&existingID)
// 	if err == nil {
// 		// sku_instance_id already exists, return the existing sku_id
// 		return existingID, nil
// 	} else if err != sql.ErrNoRows {
// 		// An error occurred during the query
// 		log.Fatalf("Error checking SKU existence: %v", err)
// 		return -1, err
// 	}

//		// If sku_instance_id doesn't exist, proceed with insertion
//		query := `INSERT INTO SKU (sku_instance_id) VALUES ($1) RETURNING sku_id`
//		var skuID int64
//		err = tx.QueryRow(query, sku_instanse_id).Scan(&skuID)
//		if err != nil {
//			log.Fatalf("Error creating SKU: %v", err)
//			return -1, err
//		}
//		return skuID, nil
//	}
func DeleteSKU(sku_instance_id string, commit bool) (bool, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		log.Errorf("Error starting transaction")
		return false, err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			if commit {
				err = tx.Commit()
			} else {
				tx.Rollback()
			}
		}

	}()

	var existingID int64
	err = tx.QueryRow("SELECT sku_id FROM SKU WHERE sku_instance_id = $1", sku_instance_id).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Errorf("SKU with sku_instance_id '%s' not found", sku_instance_id)
			return false, nil
		}
		log.Errorf("Error checking SKU existence: %v", err)
		return false, err
	}

	// If sku_instance_id exists, proceed with deletion
	query := `DELETE FROM SKU WHERE sku_instance_id = $1`
	_, err = tx.Exec(query, sku_instance_id)
	if err != nil {
		log.Fatalf("Error deleting SKU: %v", err)
		return false, err
	}

	return true, nil
}

func CreateMPO(mpo models.MPOInputParams, commit bool) (int, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		log.Errorf("Error starting transaction")
		return -1, err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			if commit {
				err = tx.Commit()
			} else {
				tx.Rollback()
			}
		}

	}()
	query := `
    INSERT INTO MPO (pdf_filename, invoice_number, mpo_instance_id)
    VALUES ($1, $2, $3)
    RETURNING mpo_id`

	var mpoID int // ID of the newly created MPO
	insert_mpo_err := tx.QueryRow(query, mpo.PDFFilename, mpo.InvoiceNumber, mpo.MPOInstanceID).Scan(&mpoID)
	fmt.Printf("created new mpo success, MPO ID: %d\n", mpoID)
	if insert_mpo_err != nil {
		return 0, fmt.Errorf("error creating MPO: %w", insert_mpo_err)
	}
	return mpoID, nil
}

func UpdateSPO(updateSPOParams models.UpdateSpoInputParams, commit bool) error {
	//Get spo using the given spo_instance_id and store the warehouseID and status in memory and then update the warehouseID and status to the new warehouse and status.
	//With the current SPO_ID get all the po_inventory rows.
	//Iterate through each po_inventory and with that sku_id, batch and current warehouse_id get the Inventory row. ---> For that particular row if the warehouse is new --> subtract that qty from the old row(check if qty matches) and then add that qty to the new row
	//else if warehouse is same --> subtract that qty(check if qty matches) and then update that same qty to the new column name

	tx, err := database.DB.Begin()
	if err != nil {
		log.Errorf("Error starting transaction")
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			if commit {
				err = tx.Commit()
			} else {
				tx.Rollback()
			}
		}

	}()

	//GET SPO using SPO Instance ID
	spoInstanceID := updateSPOParams.SPOInstanceID
	var spo models.SPO
	getspo_query := `
        SELECT spo_id, mpo_id, spo_instance_id, warehouse_id, doa, status
              FROM spo
              WHERE spo_instance_id = $1
        `
	spo_row := tx.QueryRow(getspo_query, spoInstanceID)
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
	if updateSPOParams.WareHouseID != currentWarehouse {
		newWarehouse = true
	}

	//Get Purchase Order Inventory Rows
	get_poInventory_query := `
        SELECT poi_id, sku_id, spo_id, qty, batch
              FROM po_inventory
              WHERE spo_id = $1
    `

	log.Info("SPO ID: ", spo.SPOID)
	poInventoryRows, err := tx.Query(get_poInventory_query, spo.SPOID)
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
		inventory_row := tx.QueryRow(get_inventory_query, poi.SKUID, poi.Batch, currentWarehouse)
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

		err := tx.QueryRow(get_currentValue_query, poi.SKUID, poi.Batch, currentWarehouse).Scan(&currentValue)
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
			_, err := tx.Exec(updateQuery, poi.SKUID, poi.Batch, currentWarehouse, poi.Qty)
			if err != nil {
				log.Errorf("Error Updating Inventory:")
				return err
			}
			//Add the value to the new row entry
			insertQuery := fmt.Sprintf(`
                INSERT INTO inventory (sku_id, warehouse_id, batch, %s)
                VALUES ($1, $2, $3, $4)
            `, updateSPOParams.Status)

			_, insert_err := tx.Exec(insertQuery, poi.SKUID, updateSPOParams.WareHouseID, poi.Batch, poi.Qty)
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

			_, update_inventory_err := tx.Exec(update_inventory_query, poi.SKUID, poi.Batch, currentWarehouse, poi.Qty, poi.Qty)
			if update_inventory_err != nil {
				log.Errorf("Error Updating Inventory: ", update_inventory_err)
				return update_inventory_err
			}
		}

		if updateSPOParams.Status == "in_stock" {
			// insert into transactions table for each sku
			transactionQuery := `
			INSERT INTO transactions (sku_id, spo_id, warehouse_id, bin_id, qty ,batch, type, source, expiry_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			`
			_, err = tx.Exec(transactionQuery, poi.SKUID, poi.SPOID, updateSPOParams.WareHouseID, poi.Qty, poi.Batch, "plus", "source", time.Now())
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
			_, update_spo_err := tx.Exec(update_spo_query, updateSPOParams.WareHouseID, updateSPOParams.Status, updateSPOParams.DOA, spoInstanceID)
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
func CreateMPOAndSPO(spoParams models.SPOparams, commit bool) (int, error) {

	tx, err := database.DB.Begin()
	if err != nil {
		log.Errorf("Error starting transaction")
		return -1, err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			if commit {
				err = tx.Commit()
			} else {
				tx.Rollback()
			}
		}

	}()

	// Check if mpo_id in models.SPO is present in MPO table
	fmt.Println("create mpo check")
	var mpoId int
	get_mpo_err := tx.QueryRow("SELECT mpo_id FROM MPO WHERE mpo_instance_id = $1", spoParams.Mpo.MPOInstanceID).Scan(&mpoId)
	if get_mpo_err != nil {
		_ = fmt.Errorf("error checking MPO existence: %w", get_mpo_err)
	}
	fmt.Printf("MPO ID(if 0 means no mpo): %d\n", mpoId)
	if mpoId == 0 {
		// mpo_id does not exist in MPO table, create MPO first
		newMPO := models.MPOInputParams{
			PDFFilename:   spoParams.Mpo.PDFFilename,
			InvoiceNumber: spoParams.Mpo.InvoiceNumber,
			MPOInstanceID: spoParams.Mpo.MPOInstanceID,
		}
		createdMPOID, err := CreateMPO(newMPO, true)
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
	err = tx.QueryRow(query, mpoId, spoParams.Spo.SpoInstanceId, spo_warehouse_id, spoParams.Spo.DOA, spoStatus).Scan(&spoID)
	if err != nil {
		fmt.Errorf("error creating SPO: %w", err)
	}

	// Insert into POI table
	for _, poi := range spoParams.Po_inventory {
		// get the sku_id from sku table using sku_instance_id, if not there print err
		skuID, _ := getSKUId(poi.Sku_instance_id)
		fmt.Println("SKU ID: ", skuID)
		// Insert into POI table
		query := `
			INSERT INTO po_inventory (sku_id,spo_id, qty, batch)
			VALUES ($1, $2, $3, $4)`
		_, err = tx.Exec(query, skuID, spoID, poi.Qty, poi.Batch)
		if err != nil {
			fmt.Errorf("error creating POI: %w", err)
			return -1, err
		}

		// insert into inventory table
		// check if sku_id and batch already exists in inventory table
		var invID int
		existing_inv_query := `SELECT inv_id FROM inventory WHERE sku_id = $1 AND warehouse_id = $2 AND batch = $3`
		existing_inv_err := tx.QueryRow(existing_inv_query, skuID, spo_warehouse_id, poi.Batch).Scan(&invID)
		if existing_inv_err != nil {
			if existing_inv_err == sql.ErrNoRows {
				// insert into inventory table
				insert_inv_query := `INSERT INTO inventory (sku_id, warehouse_id, batch, pending_receipt, in_stock, in_transit, received, quarantine, committed, reserved, available, damaged, bin_id)
				VALUES ($1, $2, $3, $4, 0,0,0,0,0,0,0,0,'')`
				_, insert_inv_err := tx.Exec(insert_inv_query, skuID, spoParams.Spo.WarehouseID, poi.Batch, poi.Qty)
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
        SET pending_receipt = pending_receipt + $1
        WHERE inv_id = $2`
			_, err = tx.Exec(updateQuery, poi.Qty, invID)
			if err != nil {
				log.Fatalf("Error updating inventory: %v", err)
				return -1, err
			}
		}
	}
	return spoID, nil
}

// The function `deleteSPO` deletes a record from the SPO table based on the provided SpoInstanceId.
func deleteSPO(SpoInstanceId string, commit bool) (bool, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		log.Errorf("Error starting transaction")
		return false, err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			if commit {
				err = tx.Commit()
			} else {
				tx.Rollback()
			}
		}

	}()

	query := `DELETE FROM SPO WHERE instance_id = $1`
	_, delete_spo_err := tx.Exec(query, SpoInstanceId)
	if delete_spo_err != nil {
		return false, delete_spo_err
	}
	return true, nil
}

// The function `deleteSku` deletes a SKU from the database based on the provided SKU instance ID.
func deleteSku(sku_instance_id string, commit bool) (bool, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		log.Errorf("Error starting transaction")
		return false, err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			if commit {
				err = tx.Commit()
			} else {
				tx.Rollback()
			}
		}

	}()

	query := `DELETE FROM SKU WHERE sku_instance_id = $1`
	_, delete_sku_err := tx.Exec(query, sku_instance_id)
	if delete_sku_err != nil {
		return false, delete_sku_err
	}
	fmt.Println("Deleted SKU with instance ID: ", sku_instance_id)
	return true, nil
}

func AddSPO(addSpo models.AddNewSpoInputParams, commit bool) (int, error) {

	tx, err := database.DB.Begin()
	if err != nil {
		log.Errorf("Error starting transaction")
		return -1, err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			if commit {
				err = tx.Commit()
			} else {
				tx.Rollback()
			}
		}

	}()

	// Check if mpo_id in models.SPO is present in MPO table
	fmt.Println("check addSpo")
	var mpoId int
	get_mpo_err := tx.QueryRow("SELECT mpo_id FROM MPO WHERE mpo_instance_id = $1", addSpo.MpoInstanceId).Scan(&mpoId)
	if get_mpo_err != nil {
		_ = fmt.Errorf("error checking MPO existence: %w", get_mpo_err)
	}
	fmt.Printf("MPO ID: %d\n", mpoId)
	if mpoId == 0 {
		fmt.Println("MPO with instance ID does not exist", addSpo.MpoInstanceId)
		return 0, nil
	}
	// if mpo_id exists while creating spo exists in mpo table then create SPO with the mpo_id and insert into SPO table
	query := `
			INSERT INTO SPO (mpo_id, instance_id, warehouse_id, doa, status)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING spo_id`

	var spoID int // ID of the newly created SPO
	err = tx.QueryRow(query, mpoId, addSpo.Spo.SpoInstanceId, addSpo.Spo.WarehouseID, addSpo.Spo.DOA, addSpo.Spo.Status).Scan(&spoID)
	if err != nil {
		fmt.Errorf("error creating SPO: %w", err)
	}
	//return spoID, nil

	// Insert into POI table
	for _, poi := range addSpo.Po_inventory {
		// get the sku_id from sku table using sku_instance_id, if not there print err
		skuID, _ := getSKUId(poi.Sku_instance_id)
		fmt.Println("SKU ID: ", skuID)
		// Insert into POI table
		query := ` INSERT INTO PO_Inventory (sku_id,spo_id, qty, batch)
			VALUES ($1, $2, $3, $4)`
		_, err = tx.Exec(query, skuID, spoID, poi.Qty, poi.Batch)
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
		existing_err := tx.QueryRow(existing_inv_query, skuID, addSpo.Spo.WarehouseID, poi.Batch).Scan(&invID)
		if existing_err != nil {
			if existing_err == sql.ErrNoRows {
				insertQuery := fmt.Sprintf(`
                INSERT INTO inventory (sku_id, warehouse_id, batch, %s)
                VALUES ($1, $2, $3, $4)
                RETURNING inv_id
                `, addSpo.Spo.Status)

				insert_query_err := tx.QueryRow(insertQuery, skuID, addSpo.Spo.WarehouseID, poi.Batch, poi.Qty).Scan(&invID)
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
			_, err = tx.Exec(updateQuery, poi.Qty, invID)
			if err != nil {
				log.Fatalf("Error updating inventory: %v", err)
				return -1, err
			}
		}
	}
	return spoID, nil
}

func CancelSPO(cancelSpoData models.SPO, commit bool) error {

	tx, err := database.DB.Begin()
	if err != nil {
		log.Errorf("Error starting transaction")
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			if commit {
				err = tx.Commit()
			} else {
				tx.Rollback()
			}
		}

	}()
	// get spo from spo table using spo_instance_id
	var spoID int
	var currentStatus string
	var warehouseID string
	get_spo_err := tx.QueryRow("SELECT spo_id, status, warehouse_id FROM SPO WHERE spo_instance_id = $1", cancelSpoData.SpoInstanceId).Scan(&spoID, &currentStatus, &warehouseID)
	if get_spo_err != nil {
		_ = fmt.Errorf("error checking SPO existence: %w", get_spo_err)
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
	poInventoryRows, err := tx.Query(get_poInventory_query, spoID)
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
		err := tx.QueryRow(get_currentQty_query, poi.SKUID, warehouseID, poi.Batch).Scan(&currentQty)
		if err != nil {
			return fmt.Errorf("error getting current qty from inventory: %w", err)
		}

		// subtract the qty from the inventory table with the column name same as status
		if currentQty < poi.Qty {
			return fmt.Errorf("error subtracting qty from inventory: %w", err)
		}

		query := fmt.Sprintf("UPDATE inventory SET %s = %s - $1 WHERE sku_id = $2 AND warehouse_id = $3 AND batch = $4", currentStatus, currentStatus)
		_, err = tx.Exec(query, poi.Qty, poi.SKUID, warehouseID, poi.Batch)
		if err != nil {
			return fmt.Errorf("error updating inventory: %w", err)
		}
	}

	// update the status to cancelled in spo table if not return error
	u_query := `UPDATE SPO SET status = 'cancelled' WHERE spo_id = $1`
	_, err = tx.Exec(u_query, spoID)
	if err != nil {
		return fmt.Errorf("error updating SPO: %w", err)
	}
	return nil
}

type StockingParams struct {
	SkuInstanceID  string
	SpoInstanceID  string
	Batch          string
	NewWarehouseID string
	OldWarehouseID string
	NewStatus      string
	CurrStatus     string
	BinID          string
	Qty            int
}

func StockingSKU(splitSKUParams models.StockSKUInputParams, commit bool) error {

	tx, err := database.DB.Begin()
	if err != nil {
		log.Errorf("Error starting transaction")
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			if commit {
				err = tx.Commit()
			} else {
				tx.Rollback()
			}
		}

	}()
	// dummy data for splitting sku
	splitSKUParams = models.StockSKUInputParams{
		SPOInstanceID: "sku_1",
		SKUBinParams: []models.StockSKUBinParams{
			{
				SKUInstanceID: "sku_1",
				WarehouseID:   "warehouse_1",
				BinID:         "bin_1",
				Qty:           10,
				Batch:         "batch_1",
			},
			{
				SKUInstanceID: "sku_1",
				WarehouseID:   "warehouse_2",
				BinID:         "bin_2",
				Qty:           20,
				Batch:         "batch_2",
			},
		},
	}
	// get spo
	var currentSPO models.SPO
	currentSPO, get_curr_spo_err := getSPO(splitSKUParams.SPOInstanceID)
	if get_curr_spo_err != nil {
		return get_curr_spo_err
	}

	// get po_inventory
	currPoInventories, get_po_inventory_err := getPoInventories(currentSPO.SPOID)
	if get_po_inventory_err != nil {
		return get_po_inventory_err
	}

	// get curr Po, loop and store att the old qty for sku
	skuCount := make(map[string]int)
	for _, currPoInv := range currPoInventories {
		curr_sku_string := fmt.Sprintf("%d_%s", currPoInv.SKUID, currPoInv.Batch)
		skuCount[curr_sku_string] = currPoInv.Qty
	}
	//For each SKU decrement the quantity in Inventory and Add New Transaction and Update the Inventory
	for _, skuBin := range splitSKUParams.SKUBinParams {
		//Get SkuID
		currSkuID, get_sku_err := getSKUId(skuBin.SKUInstanceID)
		if get_sku_err != nil {
			err := get_sku_err
			return err
		}

		curr_sku_string := fmt.Sprintf("%d_%s", currSkuID, skuBin.Batch)
		if skuCount[curr_sku_string] >= skuBin.Qty {
			skuCount[curr_sku_string] -= skuBin.Qty
		} else {
			err := errors.New("Invalid SKU Qty")
			return err
		}

		//Decrement The Quantity in Inventory
		//Get the current value of the column
		var currentValue int
		getCurrentValueQuery := fmt.Sprintf(`
		SELECT %s
		FROM inventory
		WHERE sku_id = $1 AND batch = $2 AND warehouse_id = $3
		`, currentSPO.Status)

		get_curr_value_err := tx.QueryRow(getCurrentValueQuery, currSkuID, skuBin.Batch, currentSPO.WarehouseID).Scan(&currentValue)
		if get_curr_value_err != nil {
			log.Errorf("Error getting current value: %v", get_curr_value_err)
			err := get_curr_value_err
			return err
		}

		if currentValue < skuBin.Qty {
			log.Errorf("Invalid Quantity")
			err := errors.New("Invalid Quantity")
			return err
		}

		//Decrement Query
		decrementQuery := fmt.Sprintf(`
		UPDATE inventory
		SET %s = %s - $4
		WHERE sku_id = $1 AND batch = $2 AND warehouse_id = $3
		`, currentSPO.Status, currentSPO.Status)

		_, decrement_err := tx.Exec(decrementQuery, currSkuID, skuBin.Batch, currentSPO.WarehouseID, skuBin.Qty)
		if decrement_err != nil {
			log.Errorf("Error Decrement Inventory Value")
			err := decrement_err
			return err
		}

		//Add Transaction
		addTransactionQuery := `
		INSERT INTO transactions (sku_id, spo_id, warehouse_id, bin_id, qty, batch, type, source, expiry_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`

		_, insertTransactionErr := tx.Exec(addTransactionQuery, currSkuID, currentSPO.SPOID, skuBin.WarehouseID, skuBin.BinID, skuBin.Qty, skuBin.Batch, "add", "Purchase Order", time.Now())
		if insertTransactionErr != nil {
			log.Errorf("Error Inserting values into transaction")
			err := insertTransactionErr
			return err
		}

		//UPDATE INVENTORY
		//Check if the row exists
		invID := 0
		checkInventoryQuery := `
		SELECT inv_id FROM inventory
		WHERE sku_id = $1 AND batch = $2 AND warehouse_id = $3 AND bin_id = $4
		`
		getInvErr := tx.QueryRow(checkInventoryQuery, currSkuID, skuBin.Batch, skuBin.WarehouseID, skuBin.BinID).Scan(&invID)
		if getInvErr != nil {
			if getInvErr == sql.ErrNoRows {
				invID = 0
			} else {
				err := getInvErr
				return err
			}
		}

		//Add New
		if invID == 0 {
			inventoryInsert_query := `
			INSERT INTO inventory (sku_id, batch, warehouse_id, bin_id, in_stock)
			VALUES ($1, $2, $3, $4, $5)
			`

			_, insertInvErr := tx.Exec(inventoryInsert_query, currSkuID, skuBin.Batch, skuBin.WarehouseID, skuBin.BinID, skuBin.Qty)
			if insertInvErr != nil {
				log.Errorf("Error inserting values into inventory")
				err := insertInvErr
				return err
			}
		} else {
			inventoryUpdate_query := `
				UPDATE inventory
				SET in_stock = COALESCE(in_stock, 0) + $1
				WHERE sku_id = $2 AND batch = $3 AND warehouse_id = $4 AND bin_id = $5
			`

			_, updateInvErr := tx.Exec(inventoryUpdate_query, skuBin.Qty, currSkuID, skuBin.Batch, skuBin.WarehouseID, skuBin.BinID)
			if updateInvErr != nil {
				log.Errorf("Error updating Inventory")
				err := updateInvErr
				return err
			}
		}

	}

	return nil
}

// func Stocking(input StockingParams) error {
// 	// get the curr qty from inventory row
// 	var currQty int
// 	get_currQty_query := fmt.Sprintf("SELECT %s FROM inventory WHERE sku_id = $1 AND warehouse_id = $2 AND batch = $3", input.CurrStatus)
// 	err := tx.QueryRow(get_currQty_query, input.SkuInstanceID, input.OldWarehouseID, input.Batch).Scan(&currQty)
// 	if err != nil {
// 		return fmt.Errorf("error getting current qty from inventory: %w", err)
// 	}

// 	// subtract the qty from the inventory table with the column name same as CurrStatus
// 	if currQty < input.Qty {
// 		return fmt.Errorf("error subtracting qty from inventory: %w", err)
// 	}

// 	query := fmt.Sprintf("UPDATE inventory SET %s = %s - $1 WHERE sku_id = $2 AND warehouse_id = $3 AND batch = $4", input.CurrStatus, input.CurrStatus)
// 	_, err = tx.Exec(query, input.Qty, input.SkuInstanceID, input.OldWarehouseID, input.Batch)
// 	if err != nil {
// 		return fmt.Errorf("error updating inventory: %w", err)
// 	}

// 	// insert the record in transactions table
// 	transactionQuery := `INSERT INTO transactions (sku_id, spo_id, warehouse_id, bin_id, qty, batch, type, source, expiry_date) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
// 	_, err = tx.Exec(transactionQuery, input.SkuInstanceID, input.SpoInstanceID, input.NewWarehouseID, input.BinID, input.Qty, input.Batch, "plus", "source", time.Now())
// 	if err != nil {
// 		return fmt.Errorf("error inserting into transactions table: %w", err)
// 	}

// 	// increament value in inventory table with in_stock status with the qty
// 	query = fmt.Sprintf("UPDATE inventory SET %s = %s + $1 WHERE sku_id = $2 AND warehouse_id = $3 AND batch = $4", input.NewStatus, input.NewStatus)
// 	_, err = tx.Exec(query, input.Qty, input.SkuInstanceID, input.NewWarehouseID, input.Batch)
// 	if err != nil {
// 		return fmt.Errorf("error updating inventory: %w", err)
// 	}

// 	// update the status to in_stock in spo table if not return error
// 	u_query := `UPDATE SPO SET status = 'in_stock' WHERE spo_instance_id = $1`
// 	_, err = tx.Exec(u_query, input.SpoInstanceID)
// 	if err != nil {
// 		return fmt.Errorf("error updating SPO: %w", err)
// 	}

// 	return nil
// }

// func to get rows in json format from the mpo table
func getMPORows() (string, error) {
	query := `
	SELECT mpo_id, pdf_filename, invoice_number, mpo_instance_id
	FROM MPO`

	mpoRows, err := database.DB.Query(query)
	if err != nil {
		log.Errorf("Error Getting MPO Rows: ", err)
		return "Error Getting MPO Rows", err
	}
	defer mpoRows.Close()

	var mpos []models.MPO
	for mpoRows.Next() {
		var mpo models.MPO
		if err := mpoRows.Scan(&mpo.MPOID, &mpo.PDFFilename, &mpo.InvoiceNumber, &mpo.MPOInstanceID); err != nil {
			log.Errorf("Error Getting MPO Row: ", err)
			return "Error Getting MPO Row", err
		}
		mpos = append(mpos, mpo)
	}
	// convert mpos to indented json
	mposJson, err := json.MarshalIndent(mpos, "", "  ")
	if err != nil {
		log.Errorf("Error converting MPO rows to JSON: ", err)
		return "Error converting MPO rows to JSON", err
	}
	// fmt.Println("MPO Rows: ", string(mposJson))
	return string(mposJson), nil
}

func getSPORows() (string, error) {
	query := `
	SELECT spo_id, mpo_id, spo_instance_id, warehouse_id, doa, status
	FROM SPO`

	spoRows, err := database.DB.Query(query)
	if err != nil {
		log.Errorf("Error Getting SPO Rows: ", err)
		return "Error Getting SPO Rows", err
	}
	defer spoRows.Close()

	var spos []models.SPO
	for spoRows.Next() {
		var spo models.SPO
		if err := spoRows.Scan(&spo.SPOID, &spo.MPOID, &spo.SpoInstanceId, &spo.WarehouseID, &spo.DOA, &spo.Status); err != nil {
			log.Errorf("Error Getting SPO Row: ", err)
			return "Error Getting SPO Row", err
		}
		spos = append(spos, spo)
	}
	// convert spos to json
	sposJson, err := json.MarshalIndent(spos, "", "  ")
	if err != nil {
		log.Errorf("Error converting SPO rows to JSON: ", err)
		return "Error converting SPO rows to JSON", err
	}
	// fmt.Println("SPO Rows: ", string(sposJson))
	return string(sposJson), nil
}

func getPOInventoryRows() (string, error) {
	query := `
	SELECT poi_id, sku_id, spo_id, qty, batch
	FROM po_inventory`

	poiRows, err := database.DB.Query(query)
	if err != nil {
		log.Errorf("Error Getting PO Inventory Rows: ", err)
		return "Error Getting PO Inventory Rows", err
	}
	defer poiRows.Close()

	var pois []models.PurchaseOrderInventory
	for poiRows.Next() {
		var poi models.PurchaseOrderInventory
		if err := poiRows.Scan(&poi.POIID, &poi.SKUID, &poi.SPOID, &poi.Qty, &poi.Batch); err != nil {
			log.Errorf("Error Getting PO Inventory Row: ", err)
			return "Error Getting PO Inventory Row", err
		}
		pois = append(pois, poi)
	}
	// convert pois to json
	poisJson, err := json.MarshalIndent(pois, "", "  ")
	if err != nil {
		log.Errorf("Error converting PO Inventory rows to JSON: ", err)
		return "Error converting PO Inventory rows to JSON", err
	}
	// fmt.Println("PO Inventory Rows: ", string(poisJson))
	return string(poisJson), nil
}

func getInventoryRows() (string, error) {
	query := `
	SELECT inv_id, sku_id, warehouse_id, batch, in_stock, pending_receipt, in_transit, received, quarantine, committed, reserved, available, damaged
	FROM inventory`

	invRows, err := database.DB.Query(query)
	if err != nil {
		log.Errorf("Error Getting Inventory Rows: ", err)
		return "Error Getting Inventory Rows", err
	}
	defer invRows.Close()

	var invs []models.Inventory
	for invRows.Next() {
		var inv models.Inventory
		if err := invRows.Scan(&inv.InvID, &inv.SKUID, &inv.WarehouseID, &inv.Batch, &inv.InStock, &inv.PendingReceipt, &inv.InTransit, &inv.Received, &inv.Quarantine, &inv.Committed, &inv.Reserved, &inv.Available, &inv.Damaged); err != nil {
			log.Errorf("Error Getting Inventory Row: ", err)
			return "Error Getting Inventory Row", err
		}
		invs = append(invs, inv)
	}
	// convert invs to json
	invsJson, err := json.MarshalIndent(invs, "", "  ")
	if err != nil {
		log.Errorf("Error converting Inventory rows to JSON: ", err)
		return "Error converting Inventory rows to JSON", err
	}
	// fmt.Println("Inventory Rows: ", string(invsJson))
	return string(invsJson), nil
}

func getSKURows() (string, error) {
	query := `
	SELECT sku_id, sku_instance_id
	FROM sku`

	skuRows, err := database.DB.Query(query)
	if err != nil {
		log.Errorf("Error Getting SKU Rows: ", err)
		return "Error Getting SKU Rows", err
	}
	defer skuRows.Close()

	var skus []models.SKU
	for skuRows.Next() {
		var sku models.SKU
		if err := skuRows.Scan(&sku.SKUID, &sku.SkuInstanceID); err != nil {
			log.Errorf("Error Getting SKU Row: ", err)
			return "Error Getting SKU Row", err
		}
		skus = append(skus, sku)
	}
	// convert skus to json
	skusJson, err := json.MarshalIndent(skus, "", "  ")
	if err != nil {
		log.Errorf("Error converting SKU rows to JSON: ", err)
		return "Error converting SKU rows to JSON", err
	}
	// fmt.Println("SKU Rows: ", string(skusJson))
	return string(skusJson), nil
}

func getTransactionsRows() (string, error) {
	query := `
	SELECT tid, sku_id, spo_id, warehouse_id, bin_id, qty, batch, type, source, expiry_date
	FROM transactions`

	transactionsRows, err := database.DB.Query(query)
	if err != nil {
		log.Errorf("Error Getting Transactions Rows: ", err)
		return "Error Getting Transactions Rows", err
	}
	defer transactionsRows.Close()

	var transactions []models.Transactions
	for transactionsRows.Next() {
		var transaction models.Transactions
		if err := transactionsRows.Scan(&transaction.TID, &transaction.SKUID, &transaction.SPOID, &transaction.WarehouseID, &transaction.BinID, &transaction.Qty, &transaction.Batch, &transaction.Type, &transaction.Source, &transaction.ExpiryDate); err != nil {
			log.Errorf("Error Getting Transactions Row: ", err)
			return "Error Getting Transactions Row", err
		}
		transactions = append(transactions, transaction)
	}
	// convert transactions to json
	transactionsJson, err := json.MarshalIndent(transactions, "", "  ")
	if err != nil {
		log.Errorf("Error converting Transactions rows to JSON: ", err)
		return "Error converting Transactions rows to JSON", err
	}
	// fmt.Println("Transactions Rows: ", string(transactionsJson))
	return string(transactionsJson), nil
}


type poisBySkuIDs struct {
	SKUID int
	POIs  []POIBySKUID
}

type POIBySKUID struct {
	POIID int
	SPOID int
	Qty   int
	Batch string
}
func getPOIBySKUID(SKUIDs[] int) (string, error) {
	get_poi_for_sku_id := `
	SELECT poi_id, spo_id, qty, batch
	FROM po_inventory
	WHERE sku_id = $1`

	var poisBySkuID []poisBySkuIDs
	for _, skuID := range SKUIDs {
		poiRows, err := database.DB.Query(get_poi_for_sku_id, skuID)
		if err != nil {
			log.Errorf("Error Getting PO Inventory Rows: ", err)
			return "Error Getting PO Inventory Rows", err
		}
		defer poiRows.Close()
		var pois []POIBySKUID
		for poiRows.Next() {
			var poi POIBySKUID
			if err := poiRows.Scan(&poi.POIID, &poi.SPOID, &poi.Qty, &poi.Batch); err != nil {
				log.Errorf("Error Getting PO Inventory Row: ", err)
				return "Error Getting PO Inventory Row", err
			}
			pois = append(pois, poi)
		}
		poisBySkuID = append(poisBySkuID, poisBySkuIDs{SKUID: skuID, POIs: pois})
	}
	// convert poisBySkuID to json
	poisBySkuIDJson, err := json.MarshalIndent(poisBySkuID, "", "  ")
	if err != nil {
		log.Errorf("Error converting PO Inventory rows to JSON: ", err)
		return "Error converting PO Inventory rows to JSON", err
	}
	// fmt.Println("PO Inventory Rows: ", string(poisBySkuIDJson))
	return string(poisBySkuIDJson), nil
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
	// 	PDFFilename:   "yo_invoice.pdf",
	// 	InvoiceNumber: "aa-11",
	// 	MPOInstanceID: "C88",
	// }
	// // Create a new MPO
	// CreateMPO(newMPO, false)

	// Create a SPO
	// newSPO := models.SPOparams{
	// 	Mpo: models.MPOInputParams{
	// 		PDFFilename:   "inv.pdf",
	// 		InvoiceNumber: "inv-123",
	// 		MPOInstanceID: "C5",
	// 	},
	// 	Spo: models.SPOInputParams{
	// 		SpoInstanceId: "SPO-1",
	// 		WarehouseID:   "W2",
	// 		DOA:           time.Now(),
	// 		Status:        "pending_receipt",
	// 	},
	// 	Po_inventory: []models.PurchaseOrderInventoryInputParams{
	// 		{
	// 			Sku_instance_id: "SKU-8",
	// 			Qty:             111,
	// 			Batch:           "BA878",
	// 		},
	// 		{
	// 			Sku_instance_id: "SKU-9",
	// 			Qty:             222,
	// 			Batch:           "BA99",
	// 		},
			
	// 	},
	// }
	// CreateMPOAndSPO(newSPO, true)

	// addNewSpoToExistingMpo := models.AddNewSpoInputParams{
	// 	MpoInstanceId: "CARPET-456",
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
	// 	Status:        "in_stock",
	// }

	// UpdateSPO(updateSPOParams)

	// stock spo
	// stockSpoData := models.StockSpoInputParams{
	// 	StockSpoArray: []models.SKUToStock{
	// 		{

	// 			Qty:         10,
	// 			WarehouseID: "W1",
	// 			BinID:       "B1",
	// 		},
	// 		{
	// 			Qty:         20,
	// 			WarehouseID: "W2",
	// 			BinID:       "B2",
	// 		},
	// 	}}
	// Stocking()

	// split spo
	// splitSPOData := models.SplitSPOInputParams{
	// 	MPOInstanceID:    "C1",
	// 	OldSPOInstanceID: "SPO-3",
	// 	NewSpos: []models.AddNewSpoInputParams{
	// 		{
	// 			Spo: models.SPOInputParams{

	// 				SpoInstanceId: "SPO-9",
	// 				WarehouseID:   "W1",
	// 				DOA:           time.Now(),
	// 				Status:        "pending_receipt",
	// 			},
	// 			Po_inventory: []models.PurchaseOrderInventoryInputParams{
	// 				{
	// 					Sku_instance_id: "SKU-99",
	// 					Qty:             10,
	// 					Batch:           "B1",
	// 				},
	// 				{
	// 					Sku_instance_id: "SKU-90",
	// 					Qty:             20,
	// 					Batch:           "B2",
	// 				},
	// 			},
	// 		},
	// 		{
	// 			Spo: models.SPOInputParams{
	// 				SpoInstanceId: "SPO-10",
	// 				WarehouseID:   "W1",
	// 				DOA:           time.Now(),
	// 				Status:        "pending_receipt",
	// 			},
	// 			Po_inventory: []models.PurchaseOrderInventoryInputParams{
	// 				{
	// 					Sku_instance_id: "SKU-80",
	// 					Qty:             10,
	// 					Batch:           "B1",
	// 				},
	// 				{
	// 					Sku_instance_id: "SKU-96",
	// 					Qty:             20,
	// 					Batch:           "B2",
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	// SplitSPO(splitSPOData)

	// get mpo table rows
	// var mpoRows, mpo_row_err = getMPORows()
	// if mpo_row_err != nil {
	// 	fmt.Println("Error getting MPO rows: ", mpo_row_err)
	// }
	// fmt.Println("MPO Rows: ", mpoRows)

	// get spo table rows
	// var spoRows, spo_rows_error = getSPORows()
	// if spo_rows_error != nil {
	// 	fmt.Println("Error getting SPO rows: ", spo_rows_error)
	// }
	// fmt.Println("SPO Rows: ", spoRows)

	// get po_inventory rows
	// var poInventoryRows, poInventoryRowsErr = getPOInventoryRows()
	// if poInventoryRowsErr != nil {
	// 	fmt.Println("Error getting PO Inventory rows: ", poInventoryRowsErr)
	// }
	// fmt.Println("PO Inventory Rows: ", poInventoryRows)

	// get inventory rows
	// var inventoryRows, inventoryRowsErr = getInventoryRows()
	// if inventoryRowsErr != nil {
	// 	fmt.Println("Error getting Inventory rows: ", inventoryRowsErr)
	// }
	// fmt.Println("Inventory Rows: ", inventoryRows)

	// get sku rows
	// var skuRows, skuRowsErr = getSKURows()
	// if skuRowsErr != nil {
	// 	fmt.Println("Error getting SKU rows: ", skuRowsErr)
	// }
	// fmt.Println("SKU Rows: ", skuRows)

	// get transactions rows
	// var transactionsRows, transactionsRowsErr = getTransactionsRows()
	// if transactionsRowsErr != nil {
	// 	fmt.Println("Error getting Transactions rows: ", transactionsRowsErr)
	// }
	// fmt.Println("Transactions Rows: ", transactionsRows)

	// Stocking sku

	// CreateSKU("SKU-1", true)
	// CreateSKU("SKU-2", true)
	// CreateSKU("SKU-3", true)
	// CreateSKU("SKU-4", true)
	// CreateSKU("SKU-5", true)
	// CreateSKU("SKU-6", true)
	// CreateSKU("SKU-7", true)
	// CreateSKU("SKU-8", true)
	// CreateSKU("SKU-9", true)
	// CreateSKU("SKU-10", true)
	// CreateSKU("SKU-11", true)
	// CreateSKU("SKU-12", true)
	// CreateSKU("SKU-13", true)
	// CreateSKU("SKU-14", true)
	// CreateSKU("SKU-15", true)
	// CreateSKU("SKU-16", true)
	// CreateSKU("SKU-17", true)
	// CreateSKU("SKU-18", true)
	// CreateSKU("SKU-19", true)
	// CreateSKU("SKU-20", true)
	// CreateSKU("SKU-21", true)



p,_ := getPOIBySKUID([]int{1,2})
fmt.Println(p)
}

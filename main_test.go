// test.go
package main

import (
	"database/sql"
	"testing"
	"time"
	"inventory-management-system/database"
	"inventory-management-system/models"
	"github.com/DATA-DOG/go-sqlmock"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	database.DB = db
	return db, mock
}

func TestCreateMPO(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	newMPO := models.MPOInputParams{
		PDFFilename:     "carpet456.pdf",
		InvoiceNumber:   "INV-456",
		Mpo_instance_id: "CARPET-456",
	}

	t.Run("Successful MPO creation", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO MPO").
			WithArgs(newMPO.PDFFilename, newMPO.InvoiceNumber, newMPO.Mpo_instance_id).
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := CreateMPO(newMPO)
		if err != nil {
			t.Errorf("Error was not expected while creating MPO: %s", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Database error", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO MPO").
			WithArgs(newMPO.PDFFilename, newMPO.InvoiceNumber, newMPO.Mpo_instance_id).
			WillReturnError(sql.ErrConnDone)

		_, err := CreateMPO(newMPO)
		if err == nil {
			t.Error("Expected an error, but got none")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestCreateSPO(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	spoParams := models.SPOparams{
		Mpo: models.MPOInputParams{
			Mpo_instance_id: "CARPET-456",
		},
		Spo: models.SPOInputParams{
			SpoInstanceId: "SPO-1",
			WarehouseID:   "W1",
			DOA:           time.Now(),
			Status:        "pending_receipt",
		},
		Po_inventory: []models.PurchaseOrderInventoryInputParams{
			{
				Sku_instance_id: "SKU-7",
				Qty:             300,
				Batch:           "BA17",
			},
		},
	}

	t.Run("MPO exists", func(t *testing.T) {
		mock.ExpectQuery("SELECT mpo_id FROM MPO WHERE mpo_instance_id = ?").
			WithArgs(spoParams.Mpo.Mpo_instance_id).
			WillReturnRows(sqlmock.NewRows([]string{"mpo_id"}).AddRow(1))

		mock.ExpectExec("INSERT INTO SPO").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO PO_INVENTORY").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := createSPO(spoParams)
		if err != nil {
			t.Errorf("Error was not expected while creating SPO: %s", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("MPO doesn't exist", func(t *testing.T) {
		mock.ExpectQuery("SELECT mpo_id FROM MPO WHERE mpo_instance_id = ?").
			WithArgs(spoParams.Mpo.Mpo_instance_id).
			WillReturnRows(sqlmock.NewRows([]string{"mpo_id"}))

		mock.ExpectExec("INSERT INTO MPO").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO SPO").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO PO_INVENTORY").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		_, err := createSPO(spoParams)
		if err != nil {
			t.Errorf("Error was not expected while creating SPO: %s", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestAddSpo(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	addSpoParams := models.AddNewSpoInputParams{
		MpiInstanceId: "CARPET-456",
		Spo: models.SPOInputParams{
			SpoInstanceId: "SPo",
			WarehouseID:   "W12345",
			DOA:           time.Now(),
			Status:        "Next status",
		},
		Po_inventory: []models.PurchaseOrderInventoryInputParams{
			{
				Sku_instance_id: "SKU-1",
				Qty:             20,
				Batch:           "B12345",
			},
		},
	}

	t.Run("Successful SPO addition", func(t *testing.T) {
		mock.ExpectQuery("SELECT mpo_id FROM MPO WHERE mpo_instance_id = ?").
			WithArgs(addSpoParams.MpiInstanceId).
			WillReturnRows(sqlmock.NewRows([]string{"mpo_id"}).AddRow(1))

		mock.ExpectExec("INSERT INTO SPO").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO PO_INVENTORY").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		_,err := addSpo(addSpoParams)
		if err != nil {
			t.Errorf("Error was not expected while adding SPO: %s", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("MPO not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT mpo_id FROM MPO WHERE mpo_instance_id = ?").
			WithArgs(addSpoParams.MpiInstanceId).
			WillReturnRows(sqlmock.NewRows([]string{"mpo_id"}))

		_,err := addSpo(addSpoParams)
		if err == nil {
			t.Error("Expected an error due to MPO not found, but got none")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestCancelSpo(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	cancelSpoData := models.CancleSpoInputParams{
		SpoInstanceId: "SPO-1",
		WarehouseID:   "W1",
		DOA:           time.Now(),
		Status:        "cancelled",
	}

	t.Run("Successful SPO cancellation", func(t *testing.T) {
		mock.ExpectExec("UPDATE SPO SET").
			WithArgs(cancelSpoData.Status, cancelSpoData.SpoInstanceId, cancelSpoData.WarehouseID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := CancelSPO(cancelSpoData)
		if err != nil {
			t.Errorf("Error was not expected while cancelling SPO: %s", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("SPO not found", func(t *testing.T) {
		mock.ExpectExec("UPDATE SPO SET").
			WithArgs(cancelSpoData.Status, cancelSpoData.SpoInstanceId, cancelSpoData.WarehouseID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := CancelSPO(cancelSpoData)
		if err == nil {
			t.Error("Expected an error due to SPO not found, but got none")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestUpdateSPO(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	updateSPOParams := models.UpdateSpoInputParams{
		SpoInstanceId: "SPO-4",
		WarehouseID:   "W1",
		DOA:           time.Now(),
		Status:        "in_stock",
	}

	t.Run("Successful SPO update", func(t *testing.T) {
		mock.ExpectExec("UPDATE SPO SET").
			WithArgs(updateSPOParams.Status, updateSPOParams.DOA, updateSPOParams.SpoInstanceId, updateSPOParams.WarehouseID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := UpdateSPO(updateSPOParams)
		if err != nil {
			t.Errorf("Error was not expected while updating SPO: %s", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("SPO not found", func(t *testing.T) {
		mock.ExpectExec("UPDATE SPO SET").
			WithArgs(updateSPOParams.Status, updateSPOParams.DOA, updateSPOParams.SpoInstanceId, updateSPOParams.WarehouseID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := UpdateSPO(updateSPOParams)
		if err == nil {
			t.Error("Expected an error due to SPO not found, but got none")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}
package main

import (
	"database/sql"
	"testing"

	"inventory-management-system/database"
	"inventory-management-system/models"
	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver
)

// Mock database for testing
type mockDB struct {
	queryRowFunc func(query string, args ...interface{}) *sql.Row
}

func (m *mockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.queryRowFunc(query, args...)
}

// TestCreateMPO tests the CreateMPO function
func TestCreateMPO(t *testing.T) {
	// Save the original DB and defer its restoration
	originalDB := database.DB
	defer func() { database.DB = originalDB }()

	tests := []struct {
		name    string
		mpo     models.MPOInputParams
		mockDB  mockDB
		wantID  int
		wantErr bool
	}{
		{
			name: "Successful MPO creation",
			mpo: models.MPOInputParams{
				PDFFilename:     "test.pdf",
				InvoiceNumber:   "INV001",
				Mpo_instance_id: "MPO001",
			},
			mockDB: mockDB{
				queryRowFunc: func(query string, args ...interface{}) *sql.Row {
					return sql.NewRow(1) // Simulate successful insertion with ID 1
				},
			},
			wantID:  1,
			wantErr: false,
		},
		{
			name: "Database error",
			mpo: models.MPOInputParams{
				PDFFilename:     "test.pdf",
				InvoiceNumber:   "INV002",
				Mpo_instance_id: "MPO002",
			},
			mockDB: mockDB{
				queryRowFunc: func(query string, args ...interface{}) *sql.Row {
					return sql.NewRow(sql.ErrNoRows) // Simulate a database error
				},
			},
			wantID:  0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the mock database for this test
			database.DB = &tt.mockDB

			gotID, err := CreateMPO(tt.mpo)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateMPO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotID != tt.wantID {
				t.Errorf("CreateMPO() = %v, want %v", gotID, tt.wantID)
			}
		})
	}
}
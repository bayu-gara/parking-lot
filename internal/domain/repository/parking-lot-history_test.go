package repository

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/bayu-gara/parking-lot/internal/domain/model"

	//external
	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewSQLParkingLotHistoryRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	type args struct {
		db *sql.DB
	}
	tests := []struct {
		name string
		args args
		want ParkingLotHistoryRepository
	}{
		{
			name: "Success",
			args: args{db: db},
			want: &SQLParkingLotHistoryRepo{db: db},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSQLParkingLotHistoryRepository(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSQLParkingLotHistoryRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLParkingLotHistoryRepo_Insert(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	type fields struct {
		DB *sql.DB
	}
	type args struct {
		ctx               context.Context
		parkingLotHistory model.ParkingLotHistory
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		mock    func()
		wantErr bool
	}{
		{
			name: "Failed to insert",
			fields: fields{
				DB: db,
			},
			args: args{
				ctx: context.Background(),
				parkingLotHistory: model.ParkingLotHistory{
					VehicleNumber: "B1234AA",
					SpotID:        "1-2-3",
				},
			},
			mock: func() {
				sqlMock.ExpectExec(
					"INSERT INTO parking_lot_history",
				).WithArgs("B1234AA", "1-2-3", sqlmock.AnyArg()).WillReturnError(errors.New("connection issue"))
			},
			wantErr: true,
		},
		{
			name: "Success insert",
			fields: fields{
				DB: db,
			},
			args: args{
				ctx: context.Background(),
				parkingLotHistory: model.ParkingLotHistory{
					VehicleNumber: "B1234AA",
					SpotID:        "1-2-3",
				},
			},
			mock: func() {
				sqlMock.ExpectExec(
					"INSERT INTO parking_lot_history",
				).WithArgs("B1234AA", "1-2-3", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			repo := &SQLParkingLotHistoryRepo{
				db: tt.fields.DB,
			}
			if err := repo.Insert(tt.args.ctx, tt.args.parkingLotHistory); (err != nil) != tt.wantErr {
				t.Errorf("SQLParkingLotHistoryRepo.Insert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSQLParkingLotHistoryRepo_FindLastParkingHistoryByVehicleNumber(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	type fields struct {
		DB *sql.DB
	}
	type args struct {
		ctx           context.Context
		vehicleNumber string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		mock       func()
		wantResult model.ParkingLotHistory
		wantErr    bool
	}{
		{
			name: "Error fetching data",
			fields: fields{
				DB: db,
			},
			args: args{
				ctx:           context.Background(),
				vehicleNumber: "B1234AA",
			},
			mock: func() {
				sqlMock.ExpectQuery(
					regexp.QuoteMeta("SELECT id, vehicle_number, spot_id, parking_date_time FROM parking_lot_history WHERE vehicle_number=? LIMIT 1"),
				).WithArgs("B1234AA").WillReturnError(errors.New("connection issue"))
			},
			wantResult: model.ParkingLotHistory{
				ID:              int64(0),
				VehicleNumber:   "",
				SpotID:          "",
				ParkingDateTime: int64(0),
			},
			wantErr: true,
		},
		{
			name: "Success fetching data",
			fields: fields{
				DB: db,
			},
			args: args{
				ctx:           context.Background(),
				vehicleNumber: "B1234AA",
			},
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "vehicle_number", "spot_id", "parking_date_time"}).AddRow(int64(1), "B1234AA", "1-2-3", int64(1746924004))
				sqlMock.ExpectQuery(
					regexp.QuoteMeta("SELECT id, vehicle_number, spot_id, parking_date_time FROM parking_lot_history WHERE vehicle_number=? LIMIT 1"),
				).WithArgs("B1234AA").WillReturnRows(rows)
			},
			wantResult: model.ParkingLotHistory{
				ID:              int64(1),
				VehicleNumber:   "B1234AA",
				SpotID:          "1-2-3",
				ParkingDateTime: int64(1746924004),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			repo := &SQLParkingLotHistoryRepo{
				db: tt.fields.DB,
			}
			gotResult, err := repo.FindLastParkingHistoryByVehicleNumber(tt.args.ctx, tt.args.vehicleNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLParkingLotHistoryRepo.FindLastParkingHistoryByVehicleNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("SQLParkingLotHistoryRepo.FindLastParkingHistoryByVehicleNumber() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

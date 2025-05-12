package repository

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/bayu-gara/parking-lot/internal/domain/model"
	redis "github.com/bayu-gara/parking-lot/pkg/redis"

	//external
	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewSQLSpotRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	type args struct {
		db          *sql.DB
		redisClient redis.Redis
	}
	tests := []struct {
		name string
		args args
		want SpotRepository
	}{
		{
			name: "Success",
			args: args{db: db},
			want: &SQLSpotRepo{db: db},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSQLSpotRepository(tt.args.db, tt.args.redisClient); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSQLSpotRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLSpotRepo_Insert(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	type fields struct {
		DB *sql.DB
	}
	type args struct {
		ctx  context.Context
		spot model.Spot
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
				spot: model.Spot{
					ID:            1002003,
					Type:          "M",
					Occupied:      false,
					VehicleNumber: "B1234AA",
					IDString:      "1-2-3",
				},
			},
			mock: func() {
				sqlMock.ExpectExec(
					"INSERT INTO spot",
				).WithArgs(1002003, "M", false, "B1234AA", "1-2-3").WillReturnError(errors.New("connection issue"))
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
				spot: model.Spot{
					ID:            1002003,
					Type:          "M",
					Occupied:      false,
					VehicleNumber: "B1234AA",
					IDString:      "1-2-3",
				},
			},
			mock: func() {
				sqlMock.ExpectExec(
					"INSERT INTO spot",
				).WithArgs(1002003, "M", false, "B1234AA", "1-2-3").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			repo := &SQLSpotRepo{
				db: tt.fields.DB,
			}
			if err := repo.Insert(tt.args.ctx, tt.args.spot); (err != nil) != tt.wantErr {
				t.Errorf("SQLSpotRepo.Insert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSQLSpotRepo_FindAvailableSpotsByVehicleType(t *testing.T) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	type fields struct {
		DB *sql.DB
	}
	type args struct {
		ctx         context.Context
		vehicleType model.VehicleType
		limit       int
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		mock      func()
		wantSpots []model.Spot
		wantErr   bool
	}{
		{
			name: "Error fetching data",
			fields: fields{
				DB: db,
			},
			args: args{
				ctx: context.Background(),
			},
			mock: func() {
				sqlMock.ExpectQuery(
					regexp.QuoteMeta("SELECT id, `type`, occupied, vehicle_number, id_string FROM spot WHERE `type`=? AND occupied=false"),
				).WillReturnError(errors.New("connection issue"))
			},
			wantSpots: nil,
			wantErr:   true,
		},
		{
			name: "Success fetching data",
			fields: fields{
				DB: db,
			},
			args: args{
				ctx: context.Background(),
			},
			mock: func() {
				rows := sqlmock.NewRows([]string{"id", "`type`", "occupied", "vehicle_number", "id_string"})
				rows.AddRow(1002003, "M", true, "B1234AA", "1-2-3")
				rows.AddRow(1002004, "B", true, "B5678CC", "1-2-4")
				sqlMock.ExpectQuery(
					regexp.QuoteMeta("SELECT id, `type`, occupied, vehicle_number, id_string FROM spot WHERE `type`=? AND occupied=false"),
				).WillReturnRows(rows)
			},
			wantSpots: []model.Spot{
				{
					ID:            1002003,
					Type:          "M",
					Occupied:      true,
					VehicleNumber: "B1234AA",
					IDString:      "1-2-3",
				},
				{
					ID:            1002004,
					Type:          "B",
					Occupied:      true,
					VehicleNumber: "B5678CC",
					IDString:      "1-2-4",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			repo := &SQLSpotRepo{
				db: tt.fields.DB,
			}
			gotSpots, err := repo.FindAvailableSpotsByVehicleType(tt.args.ctx, tt.args.vehicleType, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLSpotRepo.FindAvailableSpotByVehicleType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSpots, tt.wantSpots) {
				t.Errorf("SQLSpotRepo.FindAvailableSpotByVehicleType() = %v, want %v", gotSpots, tt.wantSpots)
			}
		})
	}
}

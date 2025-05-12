package repository

import (
	"context"
	"time"

	"github.com/bayu-gara/parking-lot/internal/domain/model"
	"github.com/bayu-gara/parking-lot/pkg/database"
)

type ParkingLotHistoryRepository interface {
	Insert(ctx context.Context, parkingLotHistory model.ParkingLotHistory) error
	FindLastParkingHistoryByVehicleNumber(ctx context.Context, vehicleNumber string) (model.ParkingLotHistory, error)
}

func NewSQLParkingLotHistoryRepository(db database.SQLDB) ParkingLotHistoryRepository {
	return &SQLParkingLotHistoryRepo{
		db: db,
	}
}

type SQLParkingLotHistoryRepo struct {
	db database.SQLDB
}

func (repo *SQLParkingLotHistoryRepo) Insert(ctx context.Context, parkingLotHistory model.ParkingLotHistory) (err error) {
	query := "INSERT INTO parking_lot_history(vehicle_number, spot_id, parking_date_time) VALUES(?,?,?)"
	_, err = repo.db.ExecContext(ctx, query, parkingLotHistory.VehicleNumber, parkingLotHistory.SpotID, time.Now().Unix())
	if err != nil {
		return err
	}

	return nil
}

func (repo *SQLParkingLotHistoryRepo) FindLastParkingHistoryByVehicleNumber(ctx context.Context, vehicleNumber string) (result model.ParkingLotHistory, err error) {
	query := "SELECT id, vehicle_number, spot_id, parking_date_time FROM parking_lot_history WHERE vehicle_number=? LIMIT 1"
	rows, err := repo.db.QueryContext(ctx, query, vehicleNumber)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&result.ID, &result.VehicleNumber, &result.SpotID, &result.ParkingDateTime)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

package repository

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/bayu-gara/parking-lot/internal/domain/model"
	"github.com/bayu-gara/parking-lot/pkg/database"
	redis "github.com/bayu-gara/parking-lot/pkg/redis"
	jsoniter "github.com/json-iterator/go"
)

type SpotRepository interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
	Insert(ctx context.Context, spot model.Spot) error
	UpsertSpot(ctx context.Context, tx *sql.Tx, spot model.Spot) error
	UpdateByOccupiedAndVehicleNumber(ctx context.Context, id int64, isOccupied bool, vehicleNumber string) error
	GetByID(ctx context.Context, id int64) (model.Spot, error)
	FindAvailableSpotsByVehicleType(ctx context.Context, vehicleType model.VehicleType, limit int) ([]model.Spot, error)
	PushAvailableSpot(ctx context.Context, spot model.Spot) error
	PopAvailableSpot(ctx context.Context, vehicleType model.VehicleType) (model.Spot, error)
}

func NewSQLSpotRepository(db database.SQLDB, redisClient redis.Redis) SpotRepository {
	return &SQLSpotRepo{
		db:          db,
		redisClient: redisClient,
	}
}

type SQLSpotRepo struct {
	db          database.SQLDB
	redisClient redis.Redis
}

func (repo *SQLSpotRepo) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return repo.db.BeginTx(ctx, nil)
}

func (repo *SQLSpotRepo) Insert(ctx context.Context, spot model.Spot) (err error) {
	query := "INSERT INTO spot(id, `type`, occupied, vehicle_number, id_string) VALUES(?,?,?,?,?)"
	_, err = repo.db.ExecContext(ctx, query, spot.ID, spot.Type, spot.Occupied, spot.VehicleNumber, spot.IDString)
	if err != nil {
		return err
	}

	return nil
}

func (repo *SQLSpotRepo) UpsertSpot(ctx context.Context, tx *sql.Tx, spot model.Spot) error {
	query := "INSERT INTO spot(id, `type`, occupied, vehicle_number, id_string) VALUES(?,?,?,?,?) "
	query = query + "ON CONFLICT (id) DO "
	query = query + "UPDATE SET `type`=?, occupied=false, vehicle_number=''"
	_, err := tx.ExecContext(ctx, query, spot.ID, spot.Type, spot.Occupied, spot.VehicleNumber, spot.IDString, spot.Type)
	if err != nil {
		return err
	}

	return nil
}

func (repo *SQLSpotRepo) UpdateByOccupiedAndVehicleNumber(ctx context.Context, id int64, isOccupied bool, vehicleNumber string) error {
	query := "UPDATE spot SET occupied=?, vehicle_number=? WHERE id=?"
	_, err := repo.db.ExecContext(ctx, query, isOccupied, vehicleNumber, id)
	if err != nil {
		return err
	}

	return nil
}

func (repo *SQLSpotRepo) GetByID(ctx context.Context, id int64) (spot model.Spot, err error) {
	query := "SELECT id, `type`, occupied, vehicle_number, id_string FROM spot WHERE id=?"
	rows, err := repo.db.QueryContext(ctx, query, id)
	if err != nil {
		return spot, err
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&spot.ID, &spot.Type, &spot.Occupied, &spot.VehicleNumber, &spot.IDString)
		if err != nil {
			return spot, err
		}
	}

	return spot, nil
}

func (repo *SQLSpotRepo) FindAvailableSpotsByVehicleType(ctx context.Context, vehicleType model.VehicleType, limit int) (spots []model.Spot, err error) {
	query := "SELECT id, `type`, occupied, vehicle_number, id_string FROM spot WHERE `type`=? AND occupied=false"
	if limit > 0 {
		query = query + " LIMIT " + strconv.FormatInt(int64(limit), 10)
	}

	rows, err := repo.db.QueryContext(ctx, query, vehicleType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var spot model.Spot
		err := rows.Scan(&spot.ID, &spot.Type, &spot.Occupied, &spot.VehicleNumber, &spot.IDString)
		if err != nil {
			return nil, err
		}
		spots = append(spots, spot)
	}

	return spots, nil
}

func (repo *SQLSpotRepo) PushAvailableSpot(ctx context.Context, spot model.Spot) error {
	redisKey := REDIS_QUEUE_AVAILABLE_PARKING_SPOT_B
	switch spot.Type {
	case "M":
		redisKey = REDIS_QUEUE_AVAILABLE_PARKING_SPOT_M
	case "A":
		redisKey = REDIS_QUEUE_AVAILABLE_PARKING_SPOT_A
	}

	jsonStr, err := jsoniter.MarshalToString(spot)
	if err != nil {
		return err
	}

	repo.redisClient.LPush(ctx, redisKey, jsonStr)

	return nil
}

func (repo *SQLSpotRepo) PopAvailableSpot(ctx context.Context, vehicleType model.VehicleType) (spot model.Spot, err error) {
	redisKey := REDIS_QUEUE_AVAILABLE_PARKING_SPOT_B
	switch vehicleType {
	case "M":
		redisKey = REDIS_QUEUE_AVAILABLE_PARKING_SPOT_M
	case "A":
		redisKey = REDIS_QUEUE_AVAILABLE_PARKING_SPOT_A
	}

	parkingSpotJsonStrRedis, err := repo.redisClient.LPop(ctx, redisKey, 1)
	if err == nil {
		if len(parkingSpotJsonStrRedis) > 0 {
			err = jsoniter.UnmarshalFromString(parkingSpotJsonStrRedis[0], &spot)
			if err == nil {
				return spot, nil
			}
		}
	}

	if len(parkingSpotJsonStrRedis) == 0 {
		parkingSpotsDB, err := repo.FindAvailableSpotsByVehicleType(ctx, vehicleType, 20)
		if err != nil {
			return spot, err
		}

		if len(parkingSpotsDB) == 1 {
			spot = parkingSpotsDB[0]
		} else if len(parkingSpotsDB) > 1 {
			jsonStrList := make([]string, 0, len(parkingSpotsDB))
			for i := 1; i < len(parkingSpotsDB); i++ {
				jsonStr, err := jsoniter.MarshalToString(parkingSpotsDB[i])
				if err != nil {
					continue
				}

				jsonStrList = append(jsonStrList, jsonStr)
			}

			repo.redisClient.RPush(ctx, redisKey, jsonStrList...)

			spot = parkingSpotsDB[0]
		}
	}

	return spot, nil
}

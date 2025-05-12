package parkinglot

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"
	"time"

	//common
	"github.com/bayu-gara/parking-lot/internal/common"

	//domain
	model "github.com/bayu-gara/parking-lot/internal/domain/model"
	repo "github.com/bayu-gara/parking-lot/internal/domain/repository"

	redis "github.com/bayu-gara/parking-lot/pkg/redis"
)

var (
	UsecaseObj Usecase
)

type Usecase interface {
	SetParkingLotMap(ctx context.Context, data [][]string) (err error)
	ParkVehicle(ctx context.Context, vehicleType model.VehicleType, vehicleNumber string) (spotID string, err error)
	UnparkVehicle(ctx context.Context, spotID string, vehicleNumber string) (err error)
	AvailableSpot(ctx context.Context, vehicleType model.VehicleType) (spotIDs []string, err error)
	SearchVehicle(ctx context.Context, vehicleNumber string) (spotID string, err error)
}

type ParkingLotUsecase struct {
	redisClient           redis.Redis
	spotRepo              repo.SpotRepository
	parkingLotHistoryRepo repo.ParkingLotHistoryRepository

	parkingLotMapLock sync.Mutex
	parkLock          sync.Mutex
}

func InitUsecase(redisClient redis.Redis, spotRepo repo.SpotRepository, parkingLotHistoryRepo repo.ParkingLotHistoryRepository) {
	UsecaseObj = &ParkingLotUsecase{
		redisClient:           redisClient,
		spotRepo:              spotRepo,
		parkingLotHistoryRepo: parkingLotHistoryRepo,
	}
}

func (p *ParkingLotUsecase) SetParkingLotMap(ctx context.Context, data [][]string) (err error) {
	if len(data) == 0 {
		return nil
	}

	// lock the request in the same app instance
	p.parkingLotMapLock.Lock()
	defer p.parkingLotMapLock.Unlock()

	ok, err := p.redisClient.SetNX(ctx, REDIS_SET_PARKING_MAP_LOCK_KEY, "1", 1800)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("System is busy, try again later")
	}

	defer p.redisClient.Delete(ctx, REDIS_SET_PARKING_MAP_LOCK_KEY)

	tx, err := p.spotRepo.BeginTx(ctx)
	if err != nil {
		return errors.New("Failed process the transaction")
	}

	for i := 1; i < len(data); i++ {
		//floor
		floor, err := strconv.Atoi(data[i][0])
		if err != nil {
			continue
		}

		//row
		row, err := strconv.Atoi(data[i][1])
		if err != nil {
			continue
		}

		//column
		column, err := strconv.Atoi(data[i][2])
		if err != nil {
			continue
		}

		//vehicle type
		vehicleType := data[i][3]

		err = p.spotRepo.UpsertSpot(ctx, tx, model.Spot{
			ID:            GenerateUniqueIDInt(floor, row, column),
			Type:          model.VehicleType(vehicleType),
			Occupied:      false,
			VehicleNumber: "",
			IDString:      GenerateUniqueIDStr(floor, row, column),
		})

		if err != nil {
			tx.Rollback()
			return errors.New("Failed to upsert parking spot data")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.New("Failed to commit upsert parking spot data")
	}

	return nil
}

func (p *ParkingLotUsecase) ParkVehicle(ctx context.Context, vehicleType model.VehicleType, vehicleNumber string) (string, error) {
	if !common.IsValidVehicleType(vehicleType) {
		return "", errors.New("Invalid vehicle type")
	}

	// lock the request in the same app instance
	p.parkLock.Lock()
	defer p.parkLock.Unlock()

	// lock the request from multiple apps
	minWait := 5
	maxWait := 10
	waitTime := rand.IntN(maxWait-minWait) + minWait
	err := p.acquireLockWithWait(ctx, REDIS_PARK_VEHICLE_LOCK_KEY, 15, waitTime)
	if err != nil {
		return "", errors.New("System busy, please try again later")
	}

	defer p.redisClient.Delete(ctx, REDIS_PARK_VEHICLE_LOCK_KEY)

	parkingSpot, err := p.spotRepo.PopAvailableSpot(ctx, vehicleType)
	if err != nil {
		return "", errors.New("Failed to get available spot")
	}

	if parkingSpot.ID <= 0 {
		return "", errors.New("no avaiable spot")
	}

	err = p.spotRepo.UpdateByOccupiedAndVehicleNumber(ctx, parkingSpot.ID, true, vehicleNumber)
	if err != nil {
		p.spotRepo.PushAvailableSpot(ctx, parkingSpot)
		return "", errors.New("Failed to reserve the spot")
	}

	p.parkingLotHistoryRepo.Insert(ctx, model.ParkingLotHistory{
		VehicleNumber: vehicleNumber,
		SpotID:        parkingSpot.IDString,
	})

	return parkingSpot.IDString, nil
}

func (p *ParkingLotUsecase) UnparkVehicle(ctx context.Context, spotID string, vehicleNumber string) (err error) {
	parkingSpot, err := p.spotRepo.GetByID(ctx, GetUniqueIDInt(spotID))
	if err != nil {
		return errors.New("parking spot with id " + spotID + " not found")
	}

	if parkingSpot.ID <= 0 {
		return errors.New("parking spot not found")
	}

	if parkingSpot.VehicleNumber != vehicleNumber {
		return errors.New("vehicle number not match")
	}

	p.spotRepo.PushAvailableSpot(ctx, parkingSpot)
	p.spotRepo.UpdateByOccupiedAndVehicleNumber(ctx, parkingSpot.ID, false, "")

	return nil
}

// ttl and maxWait unit is seconds
func (p *ParkingLotUsecase) acquireLockWithWait(ctx context.Context, key string, ttl int, maxWait int) error {
	deadline := time.Now().Add(time.Second * time.Duration(maxWait))

	for time.Now().Before(deadline) {
		ok, err := p.redisClient.SetNX(ctx, key, "1", ttl)
		if err != nil {
			return err
		}
		if ok {
			return nil // Acquired lock
		}

		time.Sleep(100 * time.Millisecond) // backoff
	}

	return fmt.Errorf("timeout waiting for lock on %s", key)
}

func (p *ParkingLotUsecase) AvailableSpot(ctx context.Context, vehicleType model.VehicleType) (spotIDs []string, err error) {
	if !common.IsValidVehicleType(vehicleType) {
		return spotIDs, errors.New("Invalid vehicle type")
	}

	spots, err := p.spotRepo.FindAvailableSpotsByVehicleType(ctx, vehicleType, 0)
	if err != nil {
		return spotIDs, errors.New("Failed to get data")
	}

	spotIDs = make([]string, 0, len(spots))
	for i := range spots {
		spotIDs = append(spotIDs, spots[i].IDString)
	}

	return spotIDs, nil
}

func (p *ParkingLotUsecase) SearchVehicle(ctx context.Context, vehicleNumber string) (string, error) {
	history, err := p.parkingLotHistoryRepo.FindLastParkingHistoryByVehicleNumber(ctx, vehicleNumber)
	if err != nil {
		return "", errors.New("Failed to get data")
	}

	return history.SpotID, nil
}

func GenerateUniqueIDInt(floor int, row int, column int) int64 {
	maxRow := 1000
	maxColumn := 1000
	return int64(floor*(maxRow*maxColumn) + row*maxColumn + column)
}

func GetUniqueIDInt(spotID string) int64 {
	spotIDSplit := strings.Split(spotID, "-")
	floor, _ := strconv.Atoi(spotIDSplit[0])
	row, _ := strconv.Atoi(spotIDSplit[1])
	column, _ := strconv.Atoi(spotIDSplit[2])

	return GenerateUniqueIDInt(floor, row, column)
}

func GenerateUniqueIDStr(floor int, row int, column int) string {
	return fmt.Sprintf("%d-%d-%d", floor, row, column)
}

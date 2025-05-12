package app

import (

	//domain

	"github.com/bayu-gara/parking-lot/internal/domain/defaults"
	repo "github.com/bayu-gara/parking-lot/internal/domain/repository"

	//usecase
	parkinglotuc "github.com/bayu-gara/parking-lot/internal/usecase/parking-lot"

	//transport
	"github.com/bayu-gara/parking-lot/internal/transport"

	//lib
	"github.com/bayu-gara/parking-lot/pkg/config"
	db "github.com/bayu-gara/parking-lot/pkg/database"
	redis "github.com/bayu-gara/parking-lot/pkg/redis"
)

func Run(mode string) error {
	cfg := config.GetConfig()

	//init database
	dbConnection, err := db.Init(cfg.Database)
	if err != nil {
		return err
	}
	defer dbConnection.Close()

	//init redis
	redisClient, err := redis.Init(cfg.Redis)
	if err != nil {
		return err
	}

	defaults.InitSQLiteTables(dbConnection)
	//defaults.GenerateDefaultSpots(dbConnection)

	spotRepo := repo.NewSQLSpotRepository(dbConnection, redisClient)
	parkinglotHistoryRepo := repo.NewSQLParkingLotHistoryRepository(dbConnection)

	//init usecase
	parkinglotuc.InitUsecase(redisClient, spotRepo, parkinglotHistoryRepo)

	server, err := transport.GetHandler(mode)
	if err != nil {
		return err
	}

	return server.Serve()
}

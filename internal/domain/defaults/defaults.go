package defaults

import (
	"fmt"
	"log"

	"github.com/bayu-gara/parking-lot/pkg/database"
)

func InitSQLiteTables(connection database.SQLDB) {
	_, err := connection.Exec(`
		CREATE TABLE IF NOT EXISTS spot (
			id INTEGER PRIMARY KEY,
			'type' TEXT,
			occupied INTEGER,
			vehicle_number TEXT,
			id_string TEXT
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = connection.Exec("CREATE INDEX IF NOT EXISTS idx_spot_vehicle_number ON spot(vehicle_number)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = connection.Exec("CREATE INDEX IF NOT EXISTS idx_spot_type_occupied_id ON spot(`type`, occupied, id ASC)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = connection.Exec(`
		CREATE TABLE IF NOT EXISTS parking_lot_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vehicle_number TEXT,
			spot_id TEXT,
			parking_date_time INTEGER
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = connection.Exec("CREATE INDEX IF NOT EXISTS idx_parking_lot_history_vehicle_number_id_desc ON parking_lot_history(vehicle_number, id DESC)")
	if err != nil {
		log.Fatal(err)
	}
}

func GenerateDefaultSpots(connection database.SQLDB) {
	var count int
	err := connection.QueryRow("SELECT COUNT(*) FROM spot").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count > 0 {
		resetDefaultSpots(connection)
		return
	}

	insertDefaultSpots(connection)
}

func insertDefaultSpots(connection database.SQLDB) {
	// Begin a transaction
	tx, err := connection.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Prepare the INSERT statement within the transaction
	query := "INSERT INTO spot(id, `type`, occupied, vehicle_number) VALUES(?,?,?,?) "
	query = query + "ON CONFLICT (id) DO "
	stmt, err := tx.Prepare("INSERT INTO spot(id, `type`, occupied, vehicle_number) VALUES(?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for floor := 1; floor <= 5; floor++ {
		for row := 1; row <= 5; row++ {
			for column := 1; column <= 5; column++ {
				spotID := fmt.Sprintf("%d-%d-%d", floor, row, column)
				vehicleType := ""

				switch column {
				case 1:
					vehicleType = "M"
				case 2, 3:
					vehicleType = "A"
				case 4:
					vehicleType = "B"
				default:
					vehicleType = "X"
				}

				_, err = stmt.Exec(spotID, vehicleType, false, "")
				if err != nil {
					tx.Rollback()
					log.Fatal(err)
				}
			}
		}
	}

	// Commit the transaction if everything executed successfully
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func resetDefaultSpots(connection database.SQLDB) {
	// Begin a transaction
	tx, err := connection.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Prepare the Update statement within the transaction
	stmt, err := tx.Prepare("UPDATE spot SET occupied = 0, vehicle_number = '' WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for floor := 1; floor <= 5; floor++ {
		for row := 1; row <= 5; row++ {
			for column := 1; column <= 5; column++ {
				spotID := fmt.Sprintf("%d-%d-%d", floor, row, column)

				_, err = stmt.Exec(spotID)
				if err != nil {
					tx.Rollback()
					log.Fatal(err)
				}
			}
		}
	}

	// Commit the transaction if everything executed successfully
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

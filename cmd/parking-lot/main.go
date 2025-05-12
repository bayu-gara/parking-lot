package main

import (
	"flag"
	"log"

	//internal
	"github.com/bayu-gara/parking-lot/internal/app"
)

func main() {
	var appMode = flag.String("mode", "rest", "app transport mode (rest, grpc, etc)")

	err := app.Run(*appMode)
	if err != nil {
		log.Fatalln("There is an error: ", err)
	}
}

package transport

import (
	"errors"

	//transport
	"github.com/bayu-gara/parking-lot/internal/transport/rest"

	//lib
	"github.com/bayu-gara/parking-lot/pkg/config"
)

type Server interface {
	Serve() error
}

var GetHandler = getHandlerFunc

func getHandlerFunc(mode string) (server Server, err error) {
	cfg := config.GetConfig()

	switch mode {
	case "rest":
		server = rest.RESTServer{Port: cfg.Transport.Rest.Port}
		return server, nil
	default:
		return nil, errors.New("another transport mode has not been implemented yet")
	}
}

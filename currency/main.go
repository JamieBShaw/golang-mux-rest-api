package main

import (
	"net"
	"os"

	protos "github.com/JamieBShaw/golang-mux-rest-api/currency/protos/currencypb"
	"github.com/JamieBShaw/golang-mux-rest-api/currency/server"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	// Setting default logger
	log := hclog.Default()

	// Generate default grpc server
	gs := grpc.NewServer()

	// Setting up our currency server
	cs := server.NewCurrency(log)

	//  Registering Currency server to grpc server
	protos.RegisterCurrencyServer(gs, cs)

	reflection.Register(gs)

	// Setting port location
	l, err := net.Listen("tcp", ":9092")

	if err != nil {
		log.Error("Unable to listen", "error", err)
		os.Exit(1)
	}

	// Serving server to requests
	gs.Serve(l)

}

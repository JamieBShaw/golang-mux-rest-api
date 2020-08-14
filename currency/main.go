package main

import (
	"net"
	"os"

	"github.com/JamieBShaw/golang-mux-rest-api/currency/data"
	protos "github.com/JamieBShaw/golang-mux-rest-api/currency/protos/currencypb"
	"github.com/JamieBShaw/golang-mux-rest-api/currency/server"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {

	// Setting default logger
	log := hclog.Default()

	rates, err := data.NewRates(log)
	if err != nil {
		log.Error("Unable to generate rates", "error", err)
		os.Exit(1)

	}

	// Generate default grpc server
	gs := grpc.NewServer()

	// Setting up our currency server
	cs := server.NewCurrency(rates, log)

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

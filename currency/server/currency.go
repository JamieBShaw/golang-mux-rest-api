package server

import (
	"context"
	"io"
	"time"

	"github.com/JamieBShaw/golang-mux-rest-api/currency/data"
	"github.com/JamieBShaw/golang-mux-rest-api/currency/protos/currencypb"
	protos "github.com/JamieBShaw/golang-mux-rest-api/currency/protos/currencypb"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Currency struct {
	rates         *data.ExchangeRates
	log           hclog.Logger
	subscriptions map[protos.Currency_SubscribeRatesServer][]*protos.RateRequest
	currencypb.UnimplementedCurrencyServer
}

func NewCurrency(r *data.ExchangeRates, l hclog.Logger) *Currency {

	c := &Currency{r, l, make(map[protos.Currency_SubscribeRatesServer][]*protos.RateRequest), currencypb.UnimplementedCurrencyServer{}}
	go c.handleUpates()
	return c
}

func (c *Currency) handleUpates() {
	ru := c.rates.MonitorRates(5 * time.Second)
	for range ru {
		c.log.Info("Got updated rates")

		// loop over subscribed clients
		for k, v := range c.subscriptions {

			// loop over subsribed rates
			for _, rr := range v {
				rate, err := c.rates.GetRate(rr.GetBase().String(), rr.GetDestination().String())
				if err != nil {
					c.log.Error("Unable to get updated rate", "base", rr.GetBase().String(), "destination", rr.GetDestination().String())

					err = k.Send(&protos.RateResponse{Base: rr.Base, Destination: rr.Destination, Rate: rate})
					if err != nil {
						c.log.Error("Unable to send updated rate", "base", rr.GetBase().String(), "destination", rr.GetDestination().String())
					}
				}
			}
		}
	}
}

func (c *Currency) GetRate(ctx context.Context, rr *protos.RateRequest) (*protos.RateResponse, error) {
	c.log.Info("Handle get rate", "base", rr.GetBase(), "destination", rr.GetDestination())

	if rr.Base == rr.Destination {
		err := status.Newf(
			codes.InvalidArgument,
			"Base currency %s can not be the same as the destination currency %s",
			rr.Base.String(),
			rr.Destination.String(),
		)

		err, wde := err.WithDetails(rr)
		if wde != nil {
			return nil, wde
		}

		return nil, err.Err()
	}

	rate, err := c.rates.GetRate(rr.GetBase().String(), rr.GetDestination().String())
	if err != nil {
		return nil, err
	}
	return &protos.RateResponse{Base: rr.GetBase(), Destination: rr.GetDestination(), Rate: rate}, nil
}

func (c *Currency) SubscribeRates(srs protos.Currency_SubscribeRatesServer) error {
	// Handles Client Messages
	for {
		rr, err := srs.Recv()
		if err == io.EOF {
			c.log.Info("Client has closed connection")
			return err
		}
		if err != nil {
			c.log.Error("Unable to recieve from client", "error", err)
			return err
		}
		c.log.Info("Handle client request from client", "request", rr)
		rrs, ok := c.subscriptions[srs]
		if !ok {
			rrs = []*protos.RateRequest{}
		}
		rrs = append(rrs, rr)
		c.subscriptions[srs] = rrs
	}
	return nil
}

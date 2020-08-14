package data

import (
	"context"
	"fmt"

	protos "github.com/JamieBShaw/golang-mux-rest-api/currency/protos/currencypb"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrProductNotFound is an error raised when a product can not be found in the database
var ErrProductNotFound = fmt.Errorf("Product not found")

// Product defines the structure for an API product
// swagger:model
type Product struct {
	// the id for the product
	//
	// required: false
	// min: 1
	ID int `json:"id"` // Unique identifier for the product

	// the name for this poduct
	//
	// required: true
	// max length: 255
	Name string `json:"name" validate:"required"`

	// the description for this poduct
	//
	// required: false
	// max length: 10000
	Description string `json:"description"`

	// the price for the product
	//
	// required: true
	// min: 0.01
	Price float64 `json:"price" validate:"required,gt=0"`

	// the SKU for the product
	//
	// required: false
	// pattern: [a-z]+-[a-z]+-[a-z]+
	SKU string `json:"sku" validate:"sku"`
}

// Products defines a slice of Product
type Products []*Product

type ProductsDB struct {
	currency protos.CurrencyClient
	log      hclog.Logger
	rates    map[string]float64
	client   protos.Currency_SubscribeRatesClient
}

func NewProductsDB(c protos.CurrencyClient, l hclog.Logger) *ProductsDB {
	pb := &ProductsDB{c, l, make(map[string]float64), nil}

	go pb.handleUpdates()

	return pb
}

func (p *ProductsDB) handleUpdates() {
	sub, err := p.currency.SubscribeRates(context.Background())
	if err != nil {
		p.log.Error("Cannot subcribe for rates", "error", err)
	}

	p.client = sub

	for {
		rr, err := sub.Recv()
		p.log.Info("Recevied updated rate from server", "dest", rr.GetDestination().String())

		if err != nil {
			p.log.Error("Error receving message", "error", err)
			return
		}
		p.rates[rr.Destination.String()] = rr.Rate
	}
}

// GetProducts returns all products from the database
func (p *ProductsDB) GetProducts(currency string) (Products, error) {

	// If currency not specified return productList
	if currency == "" {
		return productList, nil
	}

	rate, err := p.getRate(currency)
	if err != nil {
		p.log.Error("Unable to get rate", "currency", currency, "error", err)
		return nil, err
	}

	// Create new productList for response
	// Loop over productList multiplying price by response rate and append to new productList
	plr := Products{}
	for _, p := range productList {
		np := *p
		np.Price = np.Price * rate
		plr = append(plr, &np)

	}

	// return new product list
	return plr, nil

}

// GetProductByID returns a single product which matches the id from the
// database.
// If a product is not found this function returns a ProductNotFound error
func (p *ProductsDB) GetProductByID(id int, currency string) (*Product, error) {

	i := findIndexByProductID(id)
	if id == -1 {
		return nil, ErrProductNotFound
	}

	if currency == "" {
		return productList[i], nil
	}

	rate, err := p.getRate(currency)
	if err != nil {
		p.log.Error("Unable to get rate", "currency", currency, "error", err)
		return nil, err
	}
	// Need to make copy of productList as it has reference to collection, mutating it would
	// change underlining data

	np := *productList[i]
	np.Price = np.Price * rate

	return &np, nil
}

// UpdateProduct replaces a product in the database with the given
// item.
// If a product with the given id does not exist in the database
// this function returns a ProductNotFound error
func (p *ProductsDB) UpdateProduct(pr *Product) error {
	i := findIndexByProductID(pr.ID)
	if i == -1 {
		return ErrProductNotFound
	}
	// update the product in the DB
	productList[i] = pr

	return nil
}

// AddProduct adds a new product to the database
func (p *ProductsDB) AddProduct(pr *Product) {
	// get the next id in sequence
	maxID := productList[len(productList)-1].ID
	pr.ID = maxID + 1

	productList = append(productList, pr)
}

// DeleteProduct deletes a product from the database
func DeleteProduct(id int) error {
	i := findIndexByProductID(id)
	if i == -1 {
		return ErrProductNotFound
	}

	productList = append(productList[:i], productList[i+1])

	return nil
}

// findIndex finds the index of a product in the database
// returns -1 when no product can be found
func findIndexByProductID(id int) int {
	for i, p := range productList {
		if p.ID == id {
			return i
		}
	}

	return -1
}

func (p *ProductsDB) getRate(destination string) (float64, error) {

	// If cached, return
	if r, ok := p.rates[destination]; ok {
		return r, nil
	}
	// Construct request with base "EUR" to destination specificed
	rr := &protos.RateRequest{
		Base:        protos.Currencies(protos.Currencies_value["EUR"]),
		Destination: protos.Currencies(protos.Currencies_value[destination]),
	}
	// get initial req rate using GetRate
	res, err := p.currency.GetRate(context.Background(), rr)
	if err != nil {
		if s, ok := status.FromError(err); ok {
			md := s.Details()[0].(*protos.RateRequest)

			if s.Code() == codes.InvalidArgument {
				return -1, fmt.Errorf("Unable to get rate from currency server, destination and base currency cannot be the same, base: %s, dest: %s", md.Base.String(), md.Destination.String())

			}
			return -1, fmt.Errorf("Unable to get rate from currency server, base: %s, dest: %s", md.Base.String(), md.Destination.String())

		}
		return -1, err

	}
	p.rates[destination] = res.Rate // update cache

	// subscribe for updates
	p.client.Send(rr)

	//	err = p.currency.SubscribeRates(context.Background(), rr)

	return res.Rate, err
}

var productList = []*Product{
	&Product{
		ID:          1,
		Name:        "Latte",
		Description: "Frothy milky coffee",
		Price:       2.45,
		SKU:         "abc323",
	},
	&Product{
		ID:          2,
		Name:        "Esspresso",
		Description: "Short and strong coffee without milk",
		Price:       1.99,
		SKU:         "fjd34",
	},
}

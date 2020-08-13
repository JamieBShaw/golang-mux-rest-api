package handlers

import (
	"context"
	"net/http"

	"github.com/JamieBShaw/golang-mux-rest-api/products-rest-api/data"
)

// swagger:route GET /products products listProducts
// Returns a list of products from the database
// responses:
//  200: productsResponse

// ListAll handles GET requests and returns all current products
func (p *Products) ListAll(rw http.ResponseWriter, r *http.Request) {
	p.l.Println("[DEBUG] get all records")

	rw.Header().Add("Content-Type", "application/json")

	// fetch the products from the datastore
	prods := data.GetProducts()

	// serialize the list to JSON
	err := data.ToJSON(prods, rw)
	if err != nil {
		http.Error(rw, "Unable to marshal json", http.StatusInternalServerError)
	}
}

// swagger:route GET /products/{id} products listSingleProduct
// Returns a single product from the database
// responses:
//  200: productsResponse
//  404: errorResponse

// ListSingle handles GET requests
func (p *Products) ListSingle(rw http.ResponseWriter, r *http.Request) {

	rw.Header().Add("Content-Type", "application/json")
	id := getProductID(r)

	p.l.Println("[DEBUG] get record id", id)

	prod, err := data.GetProductByID(id)

	switch err {
	case nil:

	case data.ErrProductNotFound:
		p.l.Println("[ERROR] fetching product", err)

		rw.WriteHeader(http.StatusNotFound)
		data.ToJSON(&GenericError{Message: err.Error()}, rw)
		return
	default:
		p.l.Println("[ERROR] fetching product", err)

		rw.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericError{Message: err.Error()}, rw)
		return
	}

	// get exchange rate
	rr := &protos.RateRequest{}
	p.cc.GetRate(context.Background())

	err = data.ToJSON(prod, rw)
	if err != nil {
		// we should never be here but log the error just incase
		p.l.Println("[ERROR] serializing product", err)
	}
}

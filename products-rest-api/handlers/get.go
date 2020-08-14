package handlers

import (
	"net/http"

	"github.com/JamieBShaw/golang-mux-rest-api/products-rest-api/data"
)

// swagger:route GET /products products listProducts
// Returns a list of products from the database
// responses:
//  200: productsResponse

// ListAll handles GET requests and returns all current products
func (p *Products) ListAll(rw http.ResponseWriter, r *http.Request) {

	rw.Header().Add("Content-Type", "application/json")

	// Extract query params from url
	cur := r.URL.Query().Get("currency")

	// fetch the products from the datastore
	prods, err := p.db.GetProducts(cur)

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericError{Message: err.Error()}, rw)
		return
	}

	// serialize the list to JSON
	err = data.ToJSON(prods, rw)
	if err != nil {
		p.l.Error("Unable to serialize product", "error", err)
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

	// Extract query params from url
	cur := r.URL.Query().Get("currency")
	id := getProductID(r)

	p.l.Debug("Get record id", "id", id)

	prod, err := p.db.GetProductByID(id, cur)

	switch err {
	case nil:

	case data.ErrProductNotFound:
		p.l.Error("[ERROR] fetching product", err)

		rw.WriteHeader(http.StatusNotFound)
		data.ToJSON(&GenericError{Message: err.Error()}, rw)
		return
	default:
		p.l.Error("[ERROR] fetching product", err)

		rw.WriteHeader(http.StatusInternalServerError)
		data.ToJSON(&GenericError{Message: err.Error()}, rw)
		return
	}

	err = data.ToJSON(prod, rw)
	if err != nil {
		// we should never be here but log the error just incase
		p.l.Error("[ERROR] serializing product", err)
	}
}

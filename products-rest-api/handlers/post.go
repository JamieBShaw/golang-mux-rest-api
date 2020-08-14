package handlers

import (
	"github.com/JamieBShaw/golang-mux-rest-api/products-rest-api/data"

	"net/http"
)

// swagger:route POST /products products createProduct
// Create a new product
//
// responses:
//	200: productResponse
//  422: errorValidation
//  501: errorResponse

// Create handles POST requests to add new products
func (p *Products) Create(rw http.ResponseWriter, r *http.Request) {
	// fetch the product from the context
	rw.Header().Add("Content-Type", "application/json")

	prod := r.Context().Value(KeyProduct{}).(*data.Product)
	p.l.Debug("Inserting Product", "debug", prod)

	p.db.AddProduct(prod)
}

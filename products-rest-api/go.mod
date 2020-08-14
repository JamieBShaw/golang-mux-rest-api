module github.com/JamieBShaw/golang-mux-rest-api/products-rest-api

go 1.14

require (
	github.com/JamieBShaw/golang-mux-rest-api/currency v0.0.0
	github.com/go-openapi/runtime v0.19.20
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/go-hclog v0.14.1
	github.com/leodido/go-urn v1.2.0 // indirect
	google.golang.org/api v0.30.0
	google.golang.org/grpc v1.31.0
)

replace github.com/JamieBShaw/golang-mux-rest-api/currency => ../currency

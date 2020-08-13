package main

import (
	"context"

	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	protos "github.com/JamieBShaw/golang-mux-rest-api/currency/protos/currencypb"
	"github.com/JamieBShaw/golang-mux-rest-api/products-rest-api/data"
	"github.com/JamieBShaw/golang-mux-rest-api/products-rest-api/handlers"

	"github.com/go-openapi/runtime/middleware"
	goHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"google.golang.org/api/transport/grpc"
)

//var bindAddress = env.String("BIND_ADDRESS", false, ":9090", "Bind address for the server")

func main() {

	l := log.New(os.Stdout, "products-api ", log.LstdFlags)
	v := data.NewValidation()

	conn, err := grpc.Dial("localhost:9092")
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	// create currency client
	cc := protos.NewCurrencyClient(conn)

	// create the handlers
	ph := handlers.NewProducts(l, v, cc)

	// create a new serve mux and register the handlers
	sm := mux.NewRouter()

	// handlers for API
	getR := sm.Methods(http.MethodGet).Subrouter()
	getR.HandleFunc("/products", ph.ListAll)
	getR.HandleFunc("/products/{id:[0-9]+}", ph.ListSingle)

	putR := sm.Methods(http.MethodPut).Subrouter()
	putR.HandleFunc("/products", ph.Update)
	putR.Use(ph.MiddlewareValidateProduct)

	postR := sm.Methods(http.MethodPost).Subrouter()
	postR.HandleFunc("/products", ph.Create)
	postR.Use(ph.MiddlewareValidateProduct)

	deleteR := sm.Methods(http.MethodDelete).Subrouter()
	deleteR.HandleFunc("/products/{id:[0-9]+}", ph.Delete)

	// documentation handlers
	opts := middleware.RedocOpts{SpecURL: "/swagger.swag.yaml"}
	sh := middleware.Redoc(opts, nil)

	getR.Handle("/docs", sh)
	getR.Handle("/swagger.swag.yaml", http.FileServer(http.Dir("./")))

	// CORS

	ch := goHandlers.CORS(goHandlers.AllowedOrigins([]string{"http://localhost:3000"}))

	// create a new server
	s := &http.Server{
		Addr:         ":9090",
		Handler:      ch(sm),
		ErrorLog:     l,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	go func() {
		l.Println("Starting server on port 9090")

		err := s.ListenAndServe()
		if err != nil {
			l.Printf("Error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	// trap sigterm or interupt and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	// Block until a signal is received.
	sig := <-c
	log.Println("Got signal:", sig)

	// gracefully shutdown the server, waiting max 30 seconds for current operations to complete
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(ctx)
}

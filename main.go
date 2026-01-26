package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"eves-diary/handlers"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    uri := os.Getenv("MONGODB_URI")

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        log.Fatal(err)
    }

    if err := client.Ping(ctx, nil); err != nil {
        log.Fatal("MongoDB connection failed:", err)
    }

    handlers.OrdersCollection = client.Database("evesdiary").Collection("orders")

    r := mux.NewRouter()
    api := r.PathPrefix("/api").Subrouter()

    api.HandleFunc("/auth/login", handlers.LoginHandler).Methods("POST")
    api.HandleFunc("/foods", handlers.GetFoods).Methods("GET")
    api.HandleFunc("/foods", handlers.CreateFoodHandler).Methods("POST")
    api.HandleFunc("/foods/{id}", handlers.UpdateFoodHandler).Methods("PUT")
    api.HandleFunc("/foods/{id}", handlers.DeleteFoodHandler).Methods("DELETE")
    api.HandleFunc("/foods/{id}/image", handlers.GetFoodImage).Methods("GET")

    api.HandleFunc("/orders", handlers.CreateOrderHandler).Methods("POST")
    api.HandleFunc("/orders", handlers.GetOrdersHandler).Methods("GET")
    api.HandleFunc("/orders/{id}/payment", handlers.UpdateOrderPaymentHandler).Methods("PUT")
    api.HandleFunc("/orders/{id}/status", handlers.UpdateOrderStatusHandler).Methods("PUT")

    log.Println("Server starting on port 4040...")
    log.Fatal(http.ListenAndServe(":4040",
        gorillaHandlers.CORS(
            gorillaHandlers.AllowedOrigins([]string{"http://localhost:3000", "http://localhost:4200", "https://eve-diary.netlify.app/"}),
            gorillaHandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
            gorillaHandlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
        )(r),
    ))
}

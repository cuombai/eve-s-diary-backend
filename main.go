package main

import (
    "log"
    "net/http"
    "os"

    "eves-diary/handlers"
    "eves-diary/database"

    gorillaHandlers "github.com/gorilla/handlers"
    "github.com/gorilla/mux"
)

func main() {
    // Use the single DB client from database.go
    database.Init() 
    handlers.OrdersCollection = database.GetCollection( "orders")
    handlers.InitFoodHandlers()
    handlers.InitLoginHandlers()

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

    // Render sets PORT automatically, so use that instead of hardcoding 4040
    port := ":8080"
    if p := os.Getenv("PORT"); p != "" {
        port = ":" + p
    }

    log.Println("Server starting on port", port)
    log.Fatal(http.ListenAndServe(port,
        gorillaHandlers.CORS(
            gorillaHandlers.AllowedOrigins([]string{
                "http://localhost:3000",
                "http://localhost:4200",
                "https://eve-diary.netlify.app/",
            }),
            gorillaHandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
            gorillaHandlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
        )(r),
    ))
}

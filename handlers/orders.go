package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "time"
    "log"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "fmt"
    "os"
    "eves-diary/models"
    "github.com/sendgrid/sendgrid-go"
    "github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Mongo collection injected from main.go or a db package
var OrdersCollection *mongo.Collection

// POST /api/orders
func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
    var order models.Order
    if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
        http.Error(w, "Invalid order payload", http.StatusBadRequest)
        return
    }

    order.ID = primitive.NewObjectID().Hex()
    order.CreatedAt = time.Now()
    if order.Status == "" {
        order.Status = "pending"
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err := OrdersCollection.InsertOne(ctx, order)
    if err != nil {
        http.Error(w, "Failed to save order", http.StatusInternalServerError)
        return
    }

    // 🔔 Send notification email with customer details
    go func() {
        if err := SendOrderNotification(order.ID, order.CustomerName, order.CustomerPhone); err != nil {
            fmt.Println("Failed to send notification:", err)
        }
    }()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}



// GET /api/orders
func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cursor, err := OrdersCollection.Find(ctx, bson.M{})
    if err != nil {
        http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)

    var orders []models.Order
    if err := cursor.All(ctx, &orders); err != nil {
        http.Error(w, "Failed to decode orders", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(orders)
}

// PUT /api/orders/{id}/payment
func UpdateOrderPaymentHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"] // string ID

    var payload struct {
        PaymentCode string `json:"paymentCode"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Invalid payload", http.StatusBadRequest)
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    update := bson.M{"$set": bson.M{"paymentCode": payload.PaymentCode}}
    res, err := OrdersCollection.UpdateOne(ctx, bson.M{"_id": id}, update)
    if err != nil {
        http.Error(w, "Failed to update payment code", http.StatusInternalServerError)
        return
    }
    if res.MatchedCount == 0 {
        http.Error(w, "Order not found", http.StatusNotFound)
        return
    }

    var updatedOrder models.Order
    if err := OrdersCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&updatedOrder); err != nil {
        http.Error(w, "Failed to fetch updated order", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(updatedOrder)
}

// PUT /api/orders/{id}/status
func UpdateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"] // string ID

    var payload struct {
        Status string `json:"status"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Invalid payload", http.StatusBadRequest)
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    update := bson.M{"$set": bson.M{"status": payload.Status}}
    res, err := OrdersCollection.UpdateOne(ctx, bson.M{"_id": id}, update)
    if err != nil {
        http.Error(w, "Failed to update status", http.StatusInternalServerError)
        return
    }
    if res.MatchedCount == 0 {
        http.Error(w, "Order not found", http.StatusNotFound)
        return
    }

    var updatedOrder models.Order
    if err := OrdersCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&updatedOrder); err != nil {
        http.Error(w, "Failed to fetch updated order", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(updatedOrder)
}


func SendOrderNotification(orderID, customerName, customerPhone string) error {
    from := mail.NewEmail("Eves Food Diary Orders", "orders@evesfooddiary.store")
    subject := "New Order Notification"
    to := mail.NewEmail("Curtis Ombai", "curtisombai@gmail.com")

    body := fmt.Sprintf(
        "A new order has been placed.\n\n"+
            "Order ID: %s\n"+
            "Customer Name: %s\n"+
            "Customer Phone Number: %s\n",
        orderID, customerName, customerPhone,
    )

    message := mail.NewSingleEmail(from, subject, to, body, body)

    apiKey := os.Getenv("SENDGRID_API_KEY") // read from environment
    if apiKey == "" {
        log.Fatal("SENDGRID_API_KEY not set")
    }
    client := sendgrid.NewSendClient(apiKey)
    response, err := client.Send(message)
    if err != nil {
        log.Printf("SendGrid error: %v", err)
        return err
    }

    log.Printf("Email sent! Status Code: %d", response.StatusCode)
    return nil
}

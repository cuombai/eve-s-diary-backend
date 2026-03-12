package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "fmt"
    "net/smtp"
    "eves-diary/models"
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

// SendOrderNotification sends an email notification when a new order is created
func SendOrderNotification(orderID string, customerName string, customerPhone string) error {
    from := "orders@evesfooddiary.store"
    password := "Jesus4Life@2026"
    to := "curtisombai@gmail.com"

    smtpHost := "mail.privateemail.com"
    smtpPort := "587"

    subject := "New Order Notification"
    body := fmt.Sprintf(
        "
            A new order has been placed.
            Order ID: %s
            Customer Name: %s
            Customer Phone Number: %s
            
        ",
        orderID, customerName, customerPhone,
    )

    message := []byte("Subject: " + subject + "\r\n" +
        "To: " + to + "\r\n" +
        "From: " + from + "\r\n" +
        "\r\n" + body)

    auth := smtp.PlainAuth("", from, password, smtpHost)

    return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
}

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
        if err := SendOrderNotification(order); err != nil {
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


func SendOrderNotification(order models.Order) error {
    from := mail.NewEmail("Eves Food Diary Orders", "orders@evesfooddiary.store")
    subject := "New Order Notification"
    to := mail.NewEmail("Eve Adhiambo", "eveadhiambo16@gmail.com")

    // Build HTML body
    itemsHTML := ""
    for _, item := range order.Items {
        itemsHTML += fmt.Sprintf(
            "<tr><td>%s</td><td>%d</td><td>%.2f</td></tr>",
            item.Item.Name, item.Quantity, item.Item.Price,
        )
    }

    paymentCode := ""
    if order.PaymentCode != nil {
        paymentCode = *order.PaymentCode
    }

    htmlBody := fmt.Sprintf(`
        <html>
        <body style="font-family: Arial, sans-serif; color: #333;">
            <h2 style="color:#2c3e50;">New Order Received</h2>
            <p><strong>Order ID:</strong> %s</p>
            <p><strong>Customer:</strong> %s (%s)</p>
            <p><strong>Status:</strong> %s</p>
            <p><strong>Payment Code:</strong> %s</p>
            <h3>Items Ordered</h3>
            <table style="border-collapse: collapse; width: 100%%;">
                <tr style="background-color:#f2f2f2;">
                    <th style="border:1px solid #ddd; padding:8px;">Item</th>
                    <th style="border:1px solid #ddd; padding:8px;">Quantity</th>
                    <th style="border:1px solid #ddd; padding:8px;">Price</th>
                </tr>
                %s
            </table>
            <p><strong>Total Price:</strong> %.2f</p>
        </body>
        </html>
    `, order.ID, order.CustomerName, order.CustomerPhone, order.Status, paymentCode, itemsHTML, order.TotalPrice)

    // Plain text fallback
    textBody := fmt.Sprintf(
        "Order ID: %s\nCustomer: %s (%s)\nStatus: %s\nPayment Code: %s\nTotal Price: %.2f\n",
        order.ID, order.CustomerName, order.CustomerPhone, order.Status, paymentCode, order.TotalPrice,
    )

    message := mail.NewSingleEmail(from, subject, to, textBody, htmlBody)

    apiKey := os.Getenv("SENDGRID_API_KEY")
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

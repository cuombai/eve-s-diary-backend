package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)



type FoodItem struct {
    ID                 primitive.ObjectID `bson:"_id,omitempty"`
    Name               string             `bson:"name"`
    Description        string             `bson:"description"`
    Price              float64            `bson:"price"`
    ImageID            primitive.ObjectID `bson:"imageId,omitempty"` // reference to GridFS file
    AvailabilityStatus string             `bson:"availabilityStatus"`
    OnSale             bool               `bson:"onSale"`
}


// import "go.mongodb.org/mongo-driver/bson/primitive"

type Order struct {
    ID            string     `bson:"_id,omitempty" json:"id"`
    Items         []CartItem `json:"items"`
    TotalPrice    float64    `json:"totalPrice"`
    CustomerName  string     `json:"customerName"`
    CustomerPhone string     `json:"customerPhone"`
    Status        string     `json:"status"`
    CreatedAt     time.Time  `json:"createdAt"`
    PaymentCode   *string    `json:"paymentCode,omitempty"`
}


// You'll also need a CartItem struct if it's nested

type CartItem struct {
    ID       string  `json:"id"`
    Quantity int     `json:"quantity"`
    Item     FoodItem `json:"item"` // reuse your FoodItem struct
}

type Admin struct {
	Email		string	`json:"email" bson:"email"`
	Password	string 	`json:"password" bson:"password"`
}
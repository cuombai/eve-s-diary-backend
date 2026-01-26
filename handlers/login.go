package handlers

import (
    "encoding/json"
    "net/http"
    "os"
    "time"

    "eves-diary/models"
    "eves-diary/database"

    "github.com/golang-jwt/jwt/v4"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.GetCollection( "admins")

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    var creds struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    var user models.Admin
    err := userCollection.FindOne(r.Context(), bson.M{"email": creds.Email}).Decode(&user)
    if err != nil || creds.Password != user.Password {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    token, err := GenerateJWT(user.Email)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func GenerateJWT(email string) (string, error) {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        secret = "dev-secret"
    }

    claims := jwt.MapClaims{
        "email": email,
        "exp":   time.Now().Add(24 * time.Hour).Unix(),
        "iat":   time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

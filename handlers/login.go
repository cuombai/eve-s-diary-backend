package handlers

import (
    "encoding/json"
    "net/http"
    "os"
    "time"
    "log"

    "eves-diary/models"
    "eves-diary/database"

    "github.com/golang-jwt/jwt/v4"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection 
func InitLoginHandlers(){
	userCollection = database.GetCollection( "admins")
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("LoginHandler called")

    var creds struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
        log.Println("Decode error:", err)
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    log.Println("Looking up user:", creds.Email)
    var user models.Admin
    err := userCollection.FindOne(r.Context(), bson.M{"email": creds.Email}).Decode(&user)
    if err != nil {
        log.Println("DB error:", err)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    if creds.Password != user.Password {
        log.Println("Password mismatch")
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    token, err := GenerateJWT(user.Email)
    if err != nil {
        log.Println("JWT error:", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    log.Println("Login successful for:", user.Email)
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

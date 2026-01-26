// --- API HANDLERS ---
package handlers

import (
    "encoding/json"
    // "encoding/base64"
    "io"
    "net/http"
    "strconv"
	"log"
    "eves-diary/database"
    "eves-diary/models"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
	 "go.mongodb.org/mongo-driver/mongo/gridfs"
)


// FoodResponse is the struct we send back to the frontend
type FoodResponse struct {
    ID                 primitive.ObjectID `json:"id"`
    Name               string             `json:"name"`
    Description        string             `json:"description"`
    Price              float64            `json:"price"`
    ImageID            primitive.ObjectID `json:"imageId,omitempty"`
    AvailabilityStatus string             `json:"availabilityStatus"`
    OnSale             bool               `json:"onSale,omitempty"`
}


var foodCollection *mongo.Collection = database.GetCollection(database.DB, "foods")
var imageBucket, _ = gridfs.NewBucket(database.GetDatabase())



// Get all food items
func GetFoods(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var foods []FoodResponse
    cursor, err := foodCollection.Find(r.Context(), bson.M{})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer cursor.Close(r.Context())

    for cursor.Next(r.Context()) {
		var food models.FoodItem
		cursor.Decode(&food)

		resp := FoodResponse{
			ID:                 food.ID,
			Name:               food.Name,
			Description:        food.Description,
			Price:              food.Price,
			AvailabilityStatus: food.AvailabilityStatus,
			OnSale:             food.OnSale,
			ImageID: food.ImageID,
		}

		if !food.ImageID.IsZero() {
			resp.ImageID = food.ImageID	
		}

		foods = append(foods, resp)
	}


    json.NewEncoder(w).Encode(foods)
}

// Create a new food item
func CreateFoodHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(10 << 20) // limit ~10MB

    file, _, err := r.FormFile("image")
	if err != nil {
		log.Println("No image uploaded:", err)
		http.Error(w, "Image upload failed", http.StatusBadRequest)
		return
	}

	defer file.Close()

	uploadStream, err := imageBucket.OpenUploadStream(r.FormValue("name"))
	if err != nil {
		http.Error(w, "Could not open GridFS upload stream", http.StatusInternalServerError)
		return
	}
	defer uploadStream.Close()

	imgBytes, _ := io.ReadAll(file)
	_, err = uploadStream.Write(imgBytes)
	if err != nil {
		http.Error(w, "Could not write to GridFS", http.StatusInternalServerError)
		return
	}

	food := models.FoodItem{
		Name:               r.FormValue("name"),
		Description:        r.FormValue("description"),
		Price:              parsePrice(r.FormValue("price")),
		ImageID:            uploadStream.FileID.(primitive.ObjectID),
		AvailabilityStatus: r.FormValue("availabilityStatus"),
		OnSale:             r.FormValue("onSale") == "true",
	}

	_, err = foodCollection.InsertOne(r.Context(), food)

    if err != nil {
        http.Error(w, "Database insert failed", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}


func GetFoodImage(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    downloadStream, err := imageBucket.OpenDownloadStream(objID)
    if err != nil {
        http.Error(w, "Image not found", http.StatusNotFound)
        return
    }
    defer downloadStream.Close()

    w.Header().Set("Content-Type", "image/jpeg") // or detect dynamically
    io.Copy(w, downloadStream)
}


// Update an existing food item
func UpdateFoodHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(10 << 20) // limit ~10MB

    vars := mux.Vars(r)
    id := vars["id"]

    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    update := bson.M{
        "name":               r.FormValue("name"),
        "description":        r.FormValue("description"),
        "price":              parsePrice(r.FormValue("price")),
        "availabilityStatus": r.FormValue("availabilityStatus"),
        "onSale":             r.FormValue("onSale") == "true",
    }

    file, _, err := r.FormFile("image")
    if err == nil {
        defer file.Close()
        imgBytes, _ := io.ReadAll(file)
        update["image"] = imgBytes
    }

    _, err = foodCollection.UpdateOne(
        r.Context(),
        bson.M{"_id": objID},
        bson.M{"$set": update},
    )
    if err != nil {
        http.Error(w, "Update failed", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

// Delete a food item
func DeleteFoodHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    _, err = foodCollection.DeleteOne(r.Context(), bson.M{"_id": objID})
    if err != nil {
        http.Error(w, "Delete failed", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// Helper to parse price safely
func parsePrice(s string) float64 {
    if s == "" {
        return 0
    }
    p, err := strconv.ParseFloat(s, 64)
    if err != nil {
        return 0
    }
    return p
}

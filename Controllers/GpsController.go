package controllers

import (
	"context"
	configs "e/Configs"
	models "e/Models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx = context.Background()
var( upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}
clients = make(map[*websocket.Conn]string)
)

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer conn.Close()

    // Initialize the current vehicle ID to an empty string
    currentVehicleID := ""

    for {
        var location models.GPS
        if err := conn.ReadJSON(&location); err != nil {
            fmt.Println("Error reading data:", err)
            return
        }

        // Check if the received vehicle ID is different from the current one
        if location.VehiculeID != currentVehicleID {
            fmt.Println("Switching to vehicle ID:", location.VehiculeID)
            currentVehicleID = location.VehiculeID
        }

        // Send real-time updates to the client only if the vehicle ID matches
        go func(vehicleID string) {
            for {
                // Check if the vehicle ID for this connection has changed
                if vehicleID != currentVehicleID {
                    fmt.Println("Stopping updates for old vehicle ID:", vehicleID)
                    return
                }

                fullData, lat, lang, err := GetLocationData(vehicleID)
                if err != nil {
                    fmt.Println("Error starting updates:", err)
                    continue
                }

                response := struct {
                    FullData  string  `json:"fullData"`
                    Latitude  float64 `json:"lat"`
                    Longitude float64 `json:"lang"`
                }{
                    FullData:  fullData,
                    Latitude:  lat,
                    Longitude: lang,
                }

                if err := conn.WriteJSON(response); err != nil {
                    fmt.Println("Error sending to client:", err)
                    return
                }

                time.Sleep(2 * time.Second)
            }
        }(currentVehicleID)
    }
}
func GetLocationData(vehicleID string) (string, float64, float64, error) {
    fullDataKey := "location:" + vehicleID
    fullData, err := configs.GetRedisClient().Get(ctx, fullDataKey).Result()
    if err != nil {
        return "", 0.0, 0.0, err
    }   
    geoLocationKey := "vehicule:location:" + vehicleID
    results, err := configs.GetRedisClient().GeoPos(ctx, "locations", geoLocationKey).Result()
    if err != nil {
        return "", 0.0, 0.0, err
    }

    // Check if there are results and extract the latitude and longitude
    var lat, lang float64
    if len(results) > 0 {
        lat = results[0].Latitude
        lang = results[0].Longitude
    }

    return fullData, lat, lang, nil
}


func StoreLocation(w http.ResponseWriter, r *http.Request) {
    var location models.GPS

    err := json.NewDecoder(r.Body).Decode(&location)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    key := "location:" + location.VehiculeID
    data, err := json.Marshal(location)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Check if the vehicle has a previous location stored
    prevLocation, err := configs.GetRedisClient().Get(ctx, key).Result()
    if err != nil && err != redis.Nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if prevLocation != "" {
        // Calculate the distance between the new and previous locations using GEODIST
        distance, err := configs.GetRedisClient().GeoDist(ctx, "GeoADDlocations", "vehicule:location:"+location.VehiculeID, key, "m").Result()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Check if the bus has moved 50 meters or more
        if distance >= 50.0 {
            // Store new location data
            if err := configs.GetRedisClient().Set(ctx, key, data, 0).Err(); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            // Update the last known location
            _, err := configs.GetRedisClient().GeoAdd(ctx, "GeoADDlocations", &redis.GeoLocation{
                Name:      "vehicule:location:" + location.VehiculeID,
                Latitude:  location.Lat,
                Longitude: location.Lang,
            }).Result()
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
        }
    } else {
        // If no previous location is found, store the new location unconditionally
        if err := configs.GetRedisClient().Set(ctx, key, data, 0).Err(); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Update the last known location
        _, err := configs.GetRedisClient().GeoAdd(ctx, "GeoADDlocations", &redis.GeoLocation{
            Name:      "vehicule:location:" + location.VehiculeID,
            Latitude:  location.Lat,
            Longitude: location.Lang,
        }).Result()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        fmt.Printf("Received POST request data1: %+v\n", location)
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "Location data stored successfully")
}


func StoreLocationwithoutCond(w http.ResponseWriter, r *http.Request) {
    var location models.GPS

    err := json.NewDecoder(r.Body).Decode(&location)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Store the full data in Redis (you can choose your own key)
    key := "location:" + location.VehiculeID

    formattedData := fmt.Sprintf(
        "{VehiculeID:%s, Lat:%f, Lang:%f, Alt:%f, Speed:%f, Bearing:%f, Acc:%f, Addr:%s, RunningTime:%s, VersionAndroid:%s}",
        location.VehiculeID,
        location.Lat,
        location.Lang,
        location.Alt,
        location.Speed,
        location.Bearing,
        location.Acc,
        location.Addr,
        location.RunningTime,
        location.VersionAndroid,
    )

    go func() {
        if err := configs.GetRedisClient().Set(ctx, key, formattedData, 0).Err(); err != nil {
            fmt.Printf("Error storing data in Redis: %s\n", err)
        }
    }()

    lat := location.Lat
    lang := location.Lang
    go func() {
        _, err := configs.GetRedisClient().GeoAdd(ctx, "locations", &redis.GeoLocation{
            Name:      "vehicule:location:" + location.VehiculeID,
            Latitude:  lat,
            Longitude: lang,
        }).Result()
        if err != nil {
            fmt.Printf("Error adding data to Redis geo set: %s\n", err)
        }
    }()

    go func() {
        client, err := mongo.Connect(ctx, options.Client().ApplyURI(configs.EnvMongoURI()))
        if err != nil {
            fmt.Printf("Error connecting to MongoDB: %s\n", err)
            return
        }
        defer client.Disconnect(ctx)

        collectionName := "vehicule_" + location.VehiculeID
        collection := client.Database("GPS").Collection(collectionName)
        _, err = collection.InsertOne(context.Background(), location)
        if err != nil {
            fmt.Printf("Error storing data in MongoDB: %s\n", err)
        }
    }()

    fmt.Printf("Received POST request data1: %+v\n", location)

    w.WriteHeader(http.StatusOK)
    fmt.Fprintln(w, "Location data stored successfully")
}
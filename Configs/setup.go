package configs

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


var redisClient *redis.Client

func ConnectDB() {
    // Connect to MongoDB
    mongoURL := EnvMongoURI()
    client, err := mongo.NewClient(options.Client().ApplyURI(mongoURL))
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Connected to MongoDB")

   
    redisURL := EnvRedisURI()
    parsedURL, err := url.Parse(redisURL)
   redisHost := parsedURL.Hostname()
    redisPort := parsedURL.Port()

    options := &redis.Options{
        Addr:     redisHost + ":" + redisPort,
        Password: "", // Add your Redis password if required
        DB:       0,
    }
    redisClient = redis.NewClient(options)
    // Test the connection to Redis
    pong, err := redisClient.Ping(ctx).Result()
    if err != nil {
        log.Fatal("Error connecting to Redis:", err)
    }
    fmt.Println("Connected to Redis:", pong)
}
// GetRedisClient returns the Redis client instance.
func GetRedisClient() *redis.Client {
    return redisClient
}
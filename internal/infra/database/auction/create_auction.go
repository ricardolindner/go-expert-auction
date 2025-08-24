package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IMongoCollection interface {
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error)
}

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection IMongoCollection
	Mutex      *sync.Mutex
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}

	result, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	if objectID, ok := result.InsertedID.(primitive.ObjectID); ok {
		auctionEntity.Id = objectID.Hex()
	}

	startAuctionCloseRoutine(ctx, ar, auctionEntityMongo)

	return nil
}

func startAuctionCloseRoutine(ctx context.Context, ar *AuctionRepository, auctionEntityMongo *AuctionEntityMongo) {
	go func() {
		duration, _ := GetAuctionDuration()
		time.Sleep(duration)
		update := bson.M{
			"$set": bson.M{
				"status":    auction_entity.Completed,
				"timestamp": time.Now().Unix(),
			},
		}
		filter := bson.M{"_id": auctionEntityMongo.Id}
		_, err := ar.Collection.UpdateOne(ctx, filter, update)
		if err != nil {
			logger.Error("Error trying to update auction status to completed", err)
			return
		}
	}()
}

func GetAuctionDuration() (time.Duration, *internal_error.InternalError) {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	if auctionInterval == "" {
		return time.Minute * 5, internal_error.NewInternalServerError("AUCTION_INTERVAL is not set, using default 5 minutes")
	}

	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return 0, internal_error.NewInternalServerError("AUCTION_INTERVAL is not a valid duration, example: 20s or 2m")
	}

	logger.Info("Auction interval set to " + auctionInterval)

	return duration, nil
}

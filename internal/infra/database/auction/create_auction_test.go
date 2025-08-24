package auction_test

import (
	"context"
	"os"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/infra/database/auction"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mockCollection struct {
	mock.Mock
}

func (m *mockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *mockCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *mockCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.SingleResult)
}

func (m *mockCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.Cursor), args.Error(1)
}

func TestCreateAuctionClosesAutomatically(t *testing.T) {
	mockCol := new(mockCollection)
	repo := &auction.AuctionRepository{
		Collection: mockCol,
	}

	auctionEntity := &auction_entity.Auction{
		Id:          "",
		ProductName: "Test Product",
		Category:    "Test Category",
		Description: "Test Desc",
		Condition:   auction_entity.New,
		Status:      auction_entity.Active,
		Timestamp:   time.Now(),
	}

	objectID := primitive.NewObjectID()

	mockCol.On("InsertOne", mock.Anything, mock.Anything).
		Return(&mongo.InsertOneResult{InsertedID: objectID}, nil)

	updateCalled := make(chan bool, 1)
	mockCol.On("UpdateOne", mock.Anything, mock.Anything, mock.Anything).
		Return(&mongo.UpdateResult{MatchedCount: 1}, nil).
		Run(func(args mock.Arguments) {
			updateCalled <- true
		})

	os.Setenv("AUCTION_INTERVAL", "100ms")
	defer os.Unsetenv("AUCTION_INTERVAL")

	err := repo.CreateAuction(context.Background(), auctionEntity)
	assert.Nil(t, err)
	assert.Equal(t, objectID.Hex(), auctionEntity.Id)

	select {
	case <-updateCalled:
	case <-time.After(1 * time.Second):
		t.Fatal("Leilão não foi fechado automaticamente")
	}

	mockCol.AssertExpectations(t)
}

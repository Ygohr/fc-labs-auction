package bid

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestFindBidByAuctionIdUsesStoredAuctionIdField(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("queries bids using auction_id", func(mt *mtest.T) {
		const auctionId = "auction-id"
		repository := &BidRepository{Collection: mt.Coll}

		mt.AddMockResponses(mtest.CreateCursorResponse(
			0,
			mt.DB.Name()+"."+mt.Coll.Name(),
			mtest.FirstBatch,
			bson.D{
				{Key: "_id", Value: "bid-id"},
				{Key: "user_id", Value: "user-id"},
				{Key: "auction_id", Value: auctionId},
				{Key: "amount", Value: 100.0},
				{Key: "timestamp", Value: int64(1)},
			},
		))

		bids, err := repository.FindBidByAuctionId(context.Background(), auctionId)
		if err != nil {
			mt.Fatalf("failed to find bids: %v", err)
		}
		if len(bids) != 1 {
			mt.Fatalf("expected one bid, got %d", len(bids))
		}

		findEvent := mt.GetStartedEvent()
		filter := findEvent.Command.Lookup("filter").Document()
		if filter.Lookup("auction_id").StringValue() != auctionId {
			mt.Fatalf("expected auction_id filter to equal %q", auctionId)
		}
	})
}

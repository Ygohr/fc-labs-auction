package auction

import (
	"context"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCreateAuctionAutomaticallyClosesAuction(t *testing.T) {
	const auctionDuration = 10 * time.Millisecond

	t.Setenv("AUCTION_DURATION", auctionDuration.String())

	insertCommands := make(chan bson.Raw, 1)
	updateCommands := make(chan bson.Raw, 1)
	updateSucceeded := make(chan struct{}, 1)
	monitor := &event.CommandMonitor{
		Started: func(_ context.Context, event *event.CommandStartedEvent) {
			command := append(bson.Raw(nil), event.Command...)
			switch event.CommandName {
			case "insert":
				insertCommands <- command
			case "update":
				updateCommands <- command
			}
		},
		Succeeded: func(_ context.Context, event *event.CommandSucceededEvent) {
			if event.CommandName == "update" {
				updateSucceeded <- struct{}{}
			}
		},
	}
	mt := mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		ClientOptions(options.Client().SetMonitor(monitor)))
	mt.Run("changes the auction status after the configured duration", func(mt *mtest.T) {
		repository := NewAuctionRepository(mt.DB)
		auction := &auction_entity.Auction{
			Id:          "auction-id",
			ProductName: "Product",
			Category:    "Category",
			Description: "Auction description",
			Condition:   auction_entity.New,
			Status:      auction_entity.Active,
			Timestamp:   time.Now(),
		}

		mt.AddMockResponses(
			mtest.CreateSuccessResponse(),
			mtest.CreateSuccessResponse(
				bson.E{Key: "n", Value: 1},
				bson.E{Key: "nModified", Value: 1},
			),
		)

		if err := repository.CreateAuction(context.Background(), auction); err != nil {
			mt.Fatalf("failed to create auction: %v", err)
		}
		if auction.Status != auction_entity.Active {
			mt.Fatalf("expected a newly created auction to be active, got %d", auction.Status)
		}
		insertCommand := <-insertCommands
		documents := insertCommand.Lookup("documents").Array()
		insertedDocuments, err := documents.Values()
		if err != nil || len(insertedDocuments) != 1 {
			mt.Fatalf("expected one inserted auction, got %d: %v", len(insertedDocuments), err)
		}
		insertedStatus := insertedDocuments[0].Document().Lookup("status").Int32()
		if auction_entity.AuctionStatus(insertedStatus) != auction_entity.Active {
			mt.Fatalf("expected inserted auction status %d, got %d", auction_entity.Active, insertedStatus)
		}

		var updateCommand bson.Raw
		select {
		case updateCommand = <-updateCommands:
		case <-time.After(time.Second):
			mt.Fatal("expected the auction status to be updated after its duration elapsed")
		}

		select {
		case <-updateSucceeded:
		case <-time.After(time.Second):
			mt.Fatal("expected the auction status update to complete")
		}

		updates, err := updateCommand.LookupErr("updates")
		if err != nil {
			mt.Fatalf("update command does not contain updates: %v", err)
		}
		updateValues, err := updates.Array().Values()
		if err != nil || len(updateValues) != 1 {
			mt.Fatalf("expected one auction update, got %d: %v", len(updateValues), err)
		}

		updateDocument := updateValues[0].Document()
		status := updateDocument.Lookup("u", "$set", "status").Int32()
		if auction_entity.AuctionStatus(status) != auction_entity.Completed {
			mt.Fatalf("expected auction status %d, got %d", auction_entity.Completed, status)
		}
	})
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firestorepb "cloud.google.com/go/firestore/apiv1/firestorepb"
	"google.golang.org/api/iterator"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cache-tool <stats|clear>")
		os.Exit(1)
	}
	command := os.Args[1]

	ctx := context.Background()
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		log.Fatal("PROJECT_ID environment variable not set")
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer client.Close()

	collectionName := "sbcwaste_cache"
	collection := client.Collection(collectionName)

	switch command {
	case "stats":
		aggQuery := collection.NewAggregationQuery().WithCount("all")
		results, err := aggQuery.Get(ctx)
		if err != nil {
			log.Fatalf("Failed to get aggregation query results: %v", err)
		}
		count, ok := results["all"]
		if !ok {
			log.Fatal("firestore: couldn't get alias for COUNT from results")
		}

		countValue := count.(*firestorepb.Value)
		fmt.Printf("Cache collection '%s' contains %d records.\n", collectionName, countValue.GetIntegerValue())
	case "clear":
		fmt.Printf("Clearing all documents from cache collection '%s'...\n", collectionName)
		iter := collection.Documents(ctx)
		numDeleted := 0
		batch := client.BulkWriter(ctx)

		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Fatalf("Failed to iterate for deletion: %v", err)
			}
			_, err = batch.Delete(doc.Ref)
			if err != nil {
				// Log the error but continue trying to delete other documents
				log.Printf("Failed to delete document %s: %v", doc.Ref.ID, err)
			}
			numDeleted++
		}
		batch.End()
		fmt.Printf("Successfully cleared %d records from the cache.\n", numDeleted)
	default:
		fmt.Printf("Unknown command: %s. Please use 'stats' or 'clear'.\n", command)
		os.Exit(1)
	}
}
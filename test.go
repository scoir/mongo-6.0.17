package main

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const docCount = 4
const fieldCount = 165

func main() {
	execute(os.Getenv("MONGO_16_URI"))
	execute(os.Getenv("MONGO_17_URI"))
}

func execute(uri string) {
	fmt.Println("connecting to", uri)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	coll := client.Database("example").Collection("example-coll")
	fmt.Println("dropping collection to ensure a clean slate")
	err = coll.Drop(context.Background())
	if err != nil {
		panic(err)
	}

	ownerId := primitive.NewObjectID()
	var docs []any
	for docIdx := range docCount {
		doc := bson.M{
			"owner": ownerId,
		}
		for i := range fieldCount {
			doc[createExampleField(i)] = docIdx + i
		}
		docs = append(docs, doc)
	}
	fmt.Println("inserting", len(docs), "documents with owner", ownerId)
	_, err = coll.InsertMany(context.Background(), docs)
	if err != nil {
		panic(err)
	}

	// This first aggregation is just under the threshold and works as expected
	runAggregation(context.Background(), coll, fieldCount-1)

	// This aggregation goes over the $group byte limit and no longer works
	runAggregation(context.Background(), coll, fieldCount)

	fmt.Println("finished")
}

func runAggregation(ctx context.Context, coll *mongo.Collection, count int) {
	fmt.Println("aggregating with", count, "fields")
	group := bson.M{
		"_id": "$owner",
	}

	for i := range count {
		field := createExampleField(i)
		group[field] = bson.M{
			"$sum": fmt.Sprintf("$%s", field),
		}
	}
	pipe := mongo.Pipeline{
		bson.D{{
			Key:   "$group",
			Value: group,
		}},
	}
	fmt.Println("running aggregation with", count, "fields")
	cursor, err := coll.Aggregate(ctx, pipe)
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)
	cnt := 0
	for cursor.Next(ctx) {
		cnt++
		var doc bson.M
		err := cursor.Decode(&doc)
		if err != nil {
			panic(err)
		}
		fmt.Println("found document", doc["_id"], "with owner", doc["owner"])
	}
	fmt.Println("found", cnt, "documents")
}

func createExampleField(i int) string {
	return fmt.Sprintf("field-with-extremely-long-name-123456789123456789-%d", i)
}

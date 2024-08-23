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
const fieldCount = 200

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

	group := bson.M{
		"_id": "$owner",
	}

	for i := range fieldCount {
		field := createExampleField(i)
		group[field] = bson.M{
			"$sum": fmt.Sprintf("$%s", field),
		}
	}
	pipe := mongo.Pipeline{
		bson.D{{
			"$group", group,
		}},
	}
	fmt.Println("aggregating")
	cursor, err := coll.Aggregate(context.Background(), pipe)
	if err != nil {
		panic(err)
	}
	defer cursor.Close(context.Background())
	cnt := 0
	for cursor.Next(context.Background()) {
		cnt++
		var doc bson.M
		err := cursor.Decode(&doc)
		if err != nil {
			panic(err)
		}
		fmt.Println("found document", doc["_id"], "with owner", doc["owner"])
		// fmt.Println(doc)
	}
	fmt.Println("found", cnt, "documents")
	fmt.Println("finished")
}

func createExampleField(i int) string {
	return fmt.Sprintf("field-with-extremely-long-name-123456789123456789-%d", i)
}

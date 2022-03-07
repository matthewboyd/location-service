package main

import (
	"context"
	"fmt"
	"github.com/matthewboyd/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

type MongoResult struct {
	ID         ID                `json:"_id" bson:"_id"`
	Geometry   models.Geometry   `json:"geometry" bson:"geometry"`
	Properties models.Properties `json:"properties" bson:"properties"`
}
type ID struct {
	Oid string `json:"$oid" bson:"$oid"`
}
type coordinates struct {
	latitude   float64
	longtitude float64
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	collection := client.Database("countries").Collection("NorthernIreland")

	coords := &coordinates{}
	coords.getCoordinates(ctx, collection, "BT61 9JG")
	collection2 := client.Database("countries").Collection("NorthernIrelandAttractions")
	coords.findLocations(ctx, collection2, 100)
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("Error connecting to the db", err)
	}
}

func (co *coordinates) getCoordinates(ctx context.Context, collection *mongo.Collection, postcode string) {
	filter := bson.D{{"properties.Postcode", postcode}}
	responseResult := &MongoResult{}
	rawBytes, err := collection.FindOne(ctx, filter).DecodeBytes()
	err = bson.Unmarshal(rawBytes, &responseResult)
	if err != nil {
		log.Fatalf("Error fetching postcode, %v", err)
	}
	log.Println(responseResult)
	co.longtitude = responseResult.Geometry.Coordinates[0]
	co.latitude = responseResult.Geometry.Coordinates[1]
}

func (co *coordinates) findLocations(ctx context.Context, collection *mongo.Collection, miles int) {
	distance := fmt.Sprintf("%f", float32(miles)*1609.34)
	var str = `{ "geometry.coordinates": { "$nearSphere": { "$geometry": { "type": "Point", "coordinates": [ ` + fmt.Sprintf("%f", co.longtitude) + `,` + fmt.Sprintf("%f", co.latitude) + `] }, "$maxDistance": ` + distance + ` } } }`
	var bdoc interface{}
	err := bson.UnmarshalExtJSON([]byte(str), true, &bdoc)
	if err != nil {
		log.Fatalf("There was an error umarshalling in findlocations: %v", err)
	}
	filter := bdoc
	test, err := collection.Find(ctx, filter)
	for test.Next(ctx) {
		var result bson.D
		if err := test.Decode(&result); err != nil {
			log.Fatal(err)
		}
		fmt.Println(result)
	}
	//log.Println(test)

	// Want to parse out the list of postcodes
}

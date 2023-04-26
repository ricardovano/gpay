package database

import (
	"context"

	_ "log"

	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mdbUser    = "apis"
	mdbPass    = "QWC7lOWuh7PFnCFM"
	mdbName    = "apis"
	mdbAddress = "apis.j5uvtfq.mongodb.net/?retryWrites=true&w=majority"

	//mongodb+srv://apis:<password>

)

func getCollection(collectionName string) (*mongo.Collection, context.Context) {

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://" + mdbName + ":" + mdbPass + "@" + mdbAddress).SetServerAPIOptions(serverAPI)

	ctx := context.Background()
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}

	collection := client.Database(mdbName).Collection(collectionName)
	return collection, ctx
}

func createRecord(collectionName string, data map[string]interface{}) (map[string]interface{}, error) {

	collection, ctx := getCollection(collectionName)
	req, err := collection.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	insertedId := req.InsertedID
	res := map[string]interface{}{

		"data": map[string]interface{}{

			"insertedId": insertedId,
		},
	}

	collection.Database().Client().Disconnect(ctx)
	return res, nil
}

func getAllRecords(collectionName string) (map[string]interface{}, error) {

	collection, ctx := getCollection(collectionName)
	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var products []bson.M
	for cur.Next(ctx) {
		var product bson.M
		if err = cur.Decode(&product); err != nil {
			collection.Database().Client().Disconnect(ctx)
			return nil, err
		}
		products = append(products, product)
	}
	defer cur.Close(ctx)

	res := map[string]interface{}{
		"data": products,
	}

	collection.Database().Client().Disconnect(ctx)
	return res, nil
}

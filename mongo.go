package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	User         string    `bson:"user"`
	Token        string    `bson:"token"`
	CreateOn     time.Time `bson:"create_on"`
	LastVerified time.Time `bson:"last_verified"`
	ExpiredTime  time.Time `bson:"expired_time"`
}

func mongo_init() {
	uri := "mongodb://" + config.MongoUsername + ":" + config.MongoPassword + "@" + config.MongoAddr
	c, err := mongo.NewClient(options.Client().ApplyURI(uri))
	mongoClient = c
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = mongoClient.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Info:Mongo session is established!!")
}

func mongoUpdateUser(client *mongo.Client, dbName, collectionName string, user User) error {
	// Get the collection
	collection := client.Database(dbName).Collection(collectionName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{"user": user.User}
	existingUser := User{}
	err := collection.FindOne(ctx, filter).Decode(&existingUser)
	if err == nil {
		// User exists, update it
		update := bson.M{
			"$set": bson.M{
				"token":         user.Token,
				"create_on":     primitive.Timestamp{T: uint32(user.CreateOn.Unix())},
				"last_verified": primitive.Timestamp{T: uint32(user.LastVerified.Unix())},
				"expired_time":  primitive.Timestamp{T: uint32(user.ExpiredTime.Unix())},
			},
		}

		_, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return err
		}
		fmt.Printf("User '%s' updated user to mongodb successfully\n", user.User)
	} else {
		// User does not exist, insert it
		_, err := collection.InsertOne(ctx, user)
		if err != nil {
			return err
		}
		fmt.Printf("User '%s' inserted user to mongodb successfully\n", user.User)
	}

	return nil
}

func findAllUsers(client *mongo.Client, dbName, collectionName string) ([]User, error) {
	// Get the collection
	collection := client.Database(dbName).Collection(collectionName)

	// Find all users
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode the results into a slice of User structs
	users := []User{}
	for cursor.Next(ctx) {
		var user User
		err := cursor.Decode(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func IsUserValidate(email string, token string) error {

	// Check if the user ID exists in the database
	collection := mongoClient.Database(config.MongoDB).Collection(config.MongoLoginCollection)
	filter := bson.M{"user": email}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var existUser User
	err := collection.FindOne(ctx, filter).Decode(&existUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return mongo.ErrNoDocuments
		}
	}
	if existUser.Token != token {
		return fmt.Errorf("User '%s' token is not valid", token)
	}

	if IsExpired(existUser.ExpiredTime) {
		return fmt.Errorf("User '%s' is expired", email)
	}
	existUser.LastVerified = time.Now().UTC()
	existUser.ExpiredTime = TimePlusSeconds(existUser.LastVerified, config.MongoLoginExpireSec)
	return nil
}

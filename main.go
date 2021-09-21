package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//Golang Struct type to represent a simple person
type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string             `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname  string             `json:"lastname,omitempty" bson:"lastname,omitempty"`
}

var client *mongo.Client

func CreatePerson(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var person Person
	json.NewDecoder(request.Body).Decode(&person)

	collection := client.Database("shanesexample").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10+time.Second)
	result, _ := collection.InsertOne(ctx, person)
	json.NewEncoder(response).Encode(result)

}

func GetSinglePerson(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var person Person
	collection := client.Database("shanesexample").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10+time.Second)
	err := collection.FindOne(ctx, Person{ID: id}).Decode(&person)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(person)
}

func GetPersons(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var people []Person
	collection := client.Database("shanesexample").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10+time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}

	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var person Person
		cursor.Decode(&person)
		people = append(people, person)

	}

	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(people)

}

func main() {
	fmt.Println("Started OK")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	client, _ = mongo.Connect(ctx, clientOptions)

	//Hooking up routers to access the endpoints using gorilla mux multiplexer
	router := mux.NewRouter()
	router.HandleFunc("/person", CreatePerson).Methods("POST")
	router.HandleFunc("/people", GetPersons).Methods("GET")

	router.HandleFunc("/singleperson/{id}", GetSinglePerson).Methods("GET")

	http.ListenAndServe(":12345", router)
}

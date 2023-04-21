package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Person struct {
	ID        string `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string `json:"firstname,omitempty" bson:"firstname,omitempty"`
	Lastname  string `json:"lastname,omitempty" bson:"lastname,omitempty"`
}

var client *mongo.Client

func CreatePersonEndPoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var person Person
	json.NewDecoder(request.Body).Decode(&person)
	collection := client.Database("thedeveloper").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	person.ID = xid.New().String()
	result, err := collection.InsertOne(ctx, person)
	if err != nil {
		log.Fatalln(err.Error())
	}

	json.NewEncoder(response).Encode(result)
	log.Println(result)
}

func GetPeopleEndPoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var people []Person
	collection := client.Database("thedeveloper").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		log.Println(err.Error())
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
		response.Write([]byte(`{ "message":"` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(people)
	log.Println(people)
}

func GetPersonEndPoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content_type", "application/json")
	params := mux.Vars(request)
	id := string(params["id"])
	var person Person
	collection := client.Database("thedeveloper").Collection("people")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, Person{ID: id}).Decode(&person)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `}`))
		return
	}
	json.NewEncoder(response).Encode(person)
}

func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb://admin:admin@127.0.0.1:27017/").SetServerAPIOptions(serverAPI)
	client, _ = mongo.Connect(ctx, opts)
	router := mux.NewRouter()
	router.HandleFunc("/person", CreatePersonEndPoint).Methods("POST")
	router.HandleFunc("/people", GetPeopleEndPoint).Methods("GET")
	router.HandleFunc("/person/{id}", GetPersonEndPoint).Methods("GET")
	http.ListenAndServe(":12345", router)
}

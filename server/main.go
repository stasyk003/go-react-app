package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Book struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title  string `json:"title" bson:"title"`
	Author string `json:"author" bson:"author"`
}

var (
	client     *mongo.Client
	collection *mongo.Collection
)

func init() {
	// Set up a connection to the MongoDB server
	// host.internal
	clientOptions := options.Client().ApplyURI("mongodb://172.17.0.1:27017")
	var err error
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set up a collection for the "books" database
	collection = client.Database("books").Collection("books")
}

func GetBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Find all documents in the "books" collection
	cur, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, "Error finding books: " + err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())
	// Create a slice of Book structs to hold the results
	var books []Book

	// Iterate through the cursor and decode each document into a Book struct
	for cur.Next(context.Background()) {
		var book Book
		err := cur.Decode(&book)
		if err != nil {
			log.Fatal(err)
		}
		books = append(books, book)
	}

	// If an error occurred during iteration, return it
	if err := cur.Err(); err != nil {
		http.Error(w, "Error iterating over books: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the books as JSON
	json.NewEncoder(w).Encode(books)
}

func searchBooks(w http.ResponseWriter, filter bson.M) {
	cur, err := collection.Find(context.Background(), filter)
	if err != nil {
		http.Error(w, "Error finding books: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(context.Background())
	var books []Book
	for cur.Next(context.Background()) {
		var book Book
		err := cur.Decode(&book)
		if err != nil {
			log.Fatal(err)
		}
		books = append(books, book)
	}
	if err := cur.Err(); err != nil {
		http.Error(w, "Error iterating over books: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(books)
}

func GetBookByTitle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	title := r.URL.Query().Get("title")
	if title == "" {
		http.Error(w, "Missing title parameter", http.StatusBadRequest)
		return
	}
	filter := bson.M{"title": title}

	searchBooks(w, filter)
}

func GetBookByAuthor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	author := r.URL.Query().Get("author")
	if author == "" {
		http.Error(w, "Missing author parameter", http.StatusBadRequest)
		return
	}
	filter := bson.M{"author": author}
	searchBooks(w, filter)
}

func CreateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var book Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if book.Title == "" || book.Author == "" {
		http.Error(w, "Missing book fields", http.StatusBadRequest)
		return
	}

	// Insert the new book document into the "books" collection
	res, err := collection.InsertOne(context.Background(), book)
	if err != nil {
		http.Error(w, "Error creating book: "+err.Error(), http.StatusInternalServerError)
		return
	}
	book.ID = res.InsertedID.(primitive.ObjectID)
	json.NewEncoder(w).Encode(book)
}

func UpdateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract book ID from URL params
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	// Find the book with the specified ID in the database
	var book Book
	err = collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&book)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	// Decode the updated book from the request body
	var updatedBook Book
	err = json.NewDecoder(r.Body).Decode(&updatedBook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the updated book
	if updatedBook.Title == "" || updatedBook.Author == "" {
		http.Error(w, "Missing book fields", http.StatusBadRequest)
		return
	}

	// Update the book in the database
	update := bson.M{
		"$set": bson.M{
			"title":  updatedBook.Title,
			"author": updatedBook.Author,
		},
	}
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": id}, update)
	if err != nil {
		http.Error(w, "Error updating book: " + err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch the updated book from the database and return it
	err = collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&book)
	if err != nil {
		http.Error(w, "Error retrieving updated book: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(book)
}

func DeleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)
    id, err := primitive.ObjectIDFromHex(params["id"])
    if err != nil {
        http.Error(w, "Invalid book ID", http.StatusBadRequest)
        return
    }

    // Delete the document with the specified ID from the "books" collection
    result, err := collection.DeleteOne(context.Background(), bson.M{"_id": id})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Check if any documents were deleted
    if result.DeletedCount == 0 {
        // Return an empty list of books if no books were found
        var books []Book
        json.NewEncoder(w).Encode(books)
        return
    }

    // Send the updated list of books as a JSON response
    cur, err := collection.Find(context.Background(), bson.M{})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer cur.Close(context.Background())

    var books []Book
    for cur.Next(context.Background()) {
        var book Book
        err := cur.Decode(&book)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        books = append(books, book)
    }
    if err := cur.Err(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(books)
}

func main() {
	router := mux.NewRouter()

	// API endpoints
	router.HandleFunc("/books", GetBooks).Methods("GET")
	router.HandleFunc("/books/search/title", GetBookByTitle).Methods("GET")
	router.HandleFunc("/books/search/author", GetBookByAuthor).Methods("GET")
	router.HandleFunc("/books", CreateBook).Methods("POST")
	router.HandleFunc("/books/{id}", UpdateBook).Methods("PUT")
	router.HandleFunc("/books/{id}", DeleteBook).Methods("DELETE")

	// Enable CORS (Cross-Origin Resource Sharing)
	handler := cors.Default().Handler(router)

	// Start server
	log.Fatal(http.ListenAndServe(":8000", handler))
}
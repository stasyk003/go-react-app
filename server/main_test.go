package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
 
    
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectDB(t *testing.T) (*mongo.Client, *mongo.Collection) {
    // Set up a MongoDB client
    clientOptions := options.Client().ApplyURI("mongodb://172.17.0.1:27017")
    client, err := mongo.Connect(context.Background(), clientOptions)
    if err != nil {
        t.Fatalf("Failed to connect to MongoDB: %v", err)
    }

    // Ping the client to ensure that the connection is valid
    err = client.Ping(context.Background(), nil)
    if err != nil {
        t.Fatalf("Failed to ping MongoDB: %v", err)
    }

    // Set up a MongoDB collection
    collection := client.Database("books").Collection("books")

    return client, collection
}

func TestGetBooks(t *testing.T) {
    client, collection := connectDB(t)
    defer client.Disconnect(context.Background())

    // Delete all documents from the collection
    _, err := collection.DeleteMany(context.Background(), bson.M{})
    assert.NoError(t, err)

    // Insert some test books into the collection
    book1 := Book{Title: "Book 1", Author: "Author 1"}
    book2 := Book{Title: "Book 2", Author: "Author 2"}
    res, err := collection.InsertMany(context.Background(), []interface{}{book1, book2})
    assert.NoError(t, err)

    // Update the expected books slice with the generated IDs
    expectedBooks := []Book{}
    for i, id := range res.InsertedIDs {
        expectedBooks = append(expectedBooks, Book{
            ID:    id.(primitive.ObjectID),
            Title: fmt.Sprintf("Book %d", i+1),
            Author: fmt.Sprintf("Author %d", i+1),
        })
    }

    // Create a request to the GET /books endpoint
    req := httptest.NewRequest("GET", "/books", nil)
    w := httptest.NewRecorder()

    // Call the getBooks handler function
    GetBooks(w, req)

    // Check the response status code
    assert.Equal(t, http.StatusOK, w.Result().StatusCode)

    // Decode the response body into a slice of books
    var books []Book
    err = json.NewDecoder(w.Body).Decode(&books)
    assert.NoError(t, err)

    // Check that the correct books were returned
    assert.Equal(t, expectedBooks, books)
}


func TestGetBookByAuthor(t *testing.T) {
	client, collection := connectDB(t)
    defer client.Disconnect(context.Background())

    // Clear the collection
    _, err := collection.DeleteMany(context.Background(), bson.M{})
    assert.NoError(t, err)

    // Insert the test book into the collection
    book := Book{Title: "Book 1", Author: "Author 1"}
    res, err := collection.InsertOne(context.Background(), book)
    assert.NoError(t, err)

    // Get the ID of the book
    id := res.InsertedID.(primitive.ObjectID)

    // Create a request to the GET /books-by-author endpoint with a query parameter
    req := httptest.NewRequest("GET", "/books-by-author?author=Author%201", nil)
    w := httptest.NewRecorder()

    // Call the getBookByAuthor handler function
    GetBookByAuthor(w, req)

    // Check the response status code
    assert.Equal(t, http.StatusOK, w.Result().StatusCode)

    // Decode the response body into a slice of books
    var books []Book
    err = json.NewDecoder(w.Body).Decode(&books)
    assert.NoError(t, err)

    // Check that the correct book was returned
    expectedBook := Book{ID: id, Title: "Book 1", Author: "Author 1"}
    assert.Equal(t, []Book{expectedBook}, books)
}

func TestGetBookByTitle(t *testing.T) {
	client, collection := connectDB(t)
    defer client.Disconnect(context.Background())

    // Clear the collection
    _, err := collection.DeleteMany(context.Background(), bson.M{})
    assert.NoError(t, err)

    // Insert the test book into the collection
	book := Book{Title: "Title 1", Author: "Author 1"}
	res, err := collection.InsertOne(context.Background(), book)
	assert.NoError(t, err)

    // Get the ID of the book
    id := res.InsertedID.(primitive.ObjectID)

    // Create a request to the GET /books-by-author endpoint with a query parameter
    req := httptest.NewRequest("GET", "/books-by-title?title=Title%201", nil)
    w := httptest.NewRecorder()

    // Call the getBookByAuthor handler function
    GetBookByTitle(w, req)

    // Check the response status code
    assert.Equal(t, http.StatusOK, w.Result().StatusCode)

    // Decode the response body into a slice of books
    var books []Book
    err = json.NewDecoder(w.Body).Decode(&books)
    assert.NoError(t, err)

    // Check that the correct book was returned
    expectedBook := Book{ID: id, Title: "Title 1", Author: "Author 1"}
    assert.Equal(t, []Book{expectedBook}, books)
}

func TestCreateBook(t *testing.T) {
    client, collection := connectDB(t)
    defer client.Disconnect(context.Background())

    // Create a new book
    newBook := Book{Title: "New Book", Author: "New Author"}

    // Encode the new book as JSON
    reqBody, err := json.Marshal(newBook)
    assert.NoError(t, err)

    // Create a request to the POST /books endpoint with the JSON-encoded new book in the request body
    req := httptest.NewRequest("POST", "/books", bytes.NewBuffer(reqBody))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    // Call the createBook handler function
    CreateBook(w, req)

    // Check the response status code
    assert.Equal(t, http.StatusOK, w.Result().StatusCode)

    // Decode the response body into the created book
    var createdBook Book
    err = json.NewDecoder(w.Body).Decode(&createdBook)
    assert.NoError(t, err)

    // Check that the created book has the correct values
    assert.NotEqual(t, primitive.NilObjectID, createdBook.ID)
    assert.Equal(t, newBook.Title, createdBook.Title)
    assert.Equal(t, newBook.Author, createdBook.Author)

    // Check that the book was actually inserted into the database
    var dbBook Book
    err = collection.FindOne(context.Background(), bson.M{"_id": createdBook.ID}).Decode(&dbBook)
    assert.NoError(t, err)
    assert.Equal(t, createdBook, dbBook)
}

func TestUpdateBook(t *testing.T) {
    // Set up a test HTTP server and client
    router := mux.NewRouter()
    router.HandleFunc("/books/{id}", UpdateBook).Methods("PUT")
    ts := httptest.NewServer(router)
    defer ts.Close()
    client := ts.Client()

    // Insert a test book into the database
    testBook := Book{
        Title:  "Test Book",
        Author: "Test Author",
    }
    result, err := collection.InsertOne(context.Background(), testBook)
    if err != nil {
        t.Fatal(err)
    }

    // Update the test book
    updatedBook := Book{
        Title:  "Updated Test Book",
        Author: "Updated Test Author",
    }
    id := result.InsertedID.(primitive.ObjectID)
    body, err := json.Marshal(updatedBook)
    if err != nil {
        t.Fatal(err)
    }
    url := ts.URL + "/books/" + id.Hex()
    req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
    if err != nil {
        t.Fatal(err)
    }
    resp, err := client.Do(req)
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", resp.StatusCode, http.StatusOK)
    }

    // Verify that the book was updated correctly
    var responseBook Book
    err = json.NewDecoder(resp.Body).Decode(&responseBook)
    if err != nil {
        t.Fatal(err)
    }
    if responseBook.Title != updatedBook.Title {
        t.Errorf("handler returned unexpected book title: got %v want %v", responseBook.Title, updatedBook.Title)
    }
    if responseBook.Author != updatedBook.Author {
        t.Errorf("handler returned unexpected book author: got %v want %v", responseBook.Author, updatedBook.Author)
    }

    // Remove the test book from the database
    _, err = collection.DeleteOne(context.Background(), bson.M{"_id": id})
    if err != nil {
        t.Fatal(err)
    }
}

func TestDeleteBook(t *testing.T) {
	// Set up a test router with the DeleteBook handler
	r := mux.NewRouter()
	r.HandleFunc("/books/{id}", DeleteBook).Methods("DELETE")

	// Create a new test HTTP request
	req, err := http.NewRequest("DELETE", "/books/6051505f5a5a430e5f771b48", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new test HTTP response recorder
	rr := httptest.NewRecorder()

	// Call the DeleteBook handler with the test request and response recorder
	r.ServeHTTP(rr, req)

	// Check that the response code is 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check that the response body is an empty array
	expected := "null\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
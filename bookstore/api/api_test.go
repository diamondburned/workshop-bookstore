package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alecthomas/assert/v2"
	"libdb.so/workshop-bookstore/bookstore"
	"libdb.so/workshop-bookstore/bookstore/db"
)

// TestBookstoreHandler tests... BookstoreHandler. The convention is to have
// TestX, where X is the name of the type or function you're testing. Go would
// love you for doing that.
func TestBookstoreHandler(t *testing.T) {
	// Make a helper function that creates a new BookstoreHandler with a
	// mockBookStorer. This way we can reuse it in multiple tests.
	newServer := func(t *testing.T, store db.Bookstore) *httptest.Server {
		server := httptest.NewServer(NewBookstoreHandler(store))
		t.Cleanup(server.Close) // stop the server when the test is done
		return server
	}

	// Define some books in stock for us to test with.
	books := []bookstore.Book{
		{
			PartialBook: bookstore.PartialBook{
				ISBN:   "978-0134190440",
				Title:  "The Go Programming Language (1st Edition)",
				Author: "Alan A. A. Donovan, Brian W. Kernighan",
				Price:  3599, // $35.99
			},
		},
		{
			PartialBook: bookstore.PartialBook{
				ISBN:   "978-1453704424",
				Title:  "The Communist Manifesto",
				Author: "Karl Marx, Friedrich Engels",
				Price:  0, // free!
			},
			Rating: ptr(4.5),
		},
	}

	// create subtests! we'll make one subtest per route that we want to
	// test on.

	t.Run("getBook", func(t *testing.T) {
		store := newMockBookstore(t, books)
		server := newServer(t, store)

		r, err := server.Client().Get(server.URL + "/books/978-0134190440")
		assert.NoError(t, err, "GET error")
		assert.Equal(t, r.StatusCode, 200, "GET status code")

		gotBook := unmarshalJSON[bookstore.Book](t, r)
		assert.Equal(t, gotBook, books[0], "GET book")
	})

	t.Run("addBook", func(t *testing.T) {
		store := newMockBookstore(t, books)
		server := newServer(t, store)

		addingBook := bookstore.Book{
			PartialBook: bookstore.PartialBook{
				ISBN:   "978-1492052593",
				Title:  "Programming Rust: Fast, Safe Systems Development (2nd Edition)",
				Author: "Jim Blandy, Jason Orendorff",
				Price:  3847, // $38.47
			},
		}

		bookJSON, err := json.Marshal(addingBook)
		assert.NoError(t, err, "marshal error")

		r, err := server.Client().Post(
			server.URL+"/books",
			"application/json", bytes.NewReader(bookJSON))
		assert.NoError(t, err, "POST error")
		assert.Equal(t, r.StatusCode, 200, "addBook POST status code")

		r2, err := server.Client().Get(server.URL + "/books/978-1492052593")
		assert.NoError(t, err, "GET error")
		assert.Equal(t, r2.StatusCode, 200, "GET status code")

		gotBook := unmarshalJSON[bookstore.Book](t, r2)
		assert.Equal(t, gotBook, addingBook, "GET book")
	})

	t.Run("getBooks", func(t *testing.T) {
		store := newMockBookstore(t, books)
		server := newServer(t, store)

		r, err := server.Client().Get(server.URL + "/books")
		assert.NoError(t, err, "GET error")
		assert.Equal(t, r.StatusCode, 200, "GET status code")

		gotBooks := unmarshalJSON[[]bookstore.PartialBook](t, r)

		expectBooks := make([]bookstore.PartialBook, len(books))
		for i, book := range books {
			expectBooks[i] = book.PartialBook
		}

		assert.Equal(t, gotBooks, expectBooks, "GET books")
	})

	t.Run("putBook", func(t *testing.T) {
		store := newMockBookstore(t, books)
		server := newServer(t, store)

		// $30.00, discounted!
		puttingBook := db.UpdateBook{Price: ptr(bookstore.Cents(3000))}

		bookJSON, err := json.Marshal(puttingBook)
		assert.NoError(t, err, "marshal error")

		req, err := http.NewRequest("PATCH",
			server.URL+"/books/978-0134190440", bytes.NewReader(bookJSON))
		assert.NoError(t, err, "NewRequest PATCH error")

		r, err := server.Client().Do(req)
		assert.NoError(t, err, "PATCH error")
		assert.Equal(t, r.StatusCode, 200, "PATCH status code")

		r2, err := server.Client().Get(server.URL + "/books/978-0134190440")
		assert.NoError(t, err, "GET error")
		assert.Equal(t, r2.StatusCode, 200, "GET status code")

		gotBook := unmarshalJSON[bookstore.Book](t, r2)
		expectBook := bookstore.Book{
			PartialBook: bookstore.PartialBook{
				ISBN:   "978-0134190440",
				Title:  "The Go Programming Language (1st Edition)",
				Author: "Alan A. A. Donovan, Brian W. Kernighan",
				Price:  3000, // $30.00
			},
		}
		assert.Equal(t, gotBook, expectBook, "GET book")
	})

	t.Run("deleteBook", func(t *testing.T) {
		store := newMockBookstore(t, books)
		server := newServer(t, store)

		req, err := http.NewRequest("DELETE", server.URL+"/books/978-0134190440", nil)
		assert.NoError(t, err, "NewRequest DELETE error")

		r, err := server.Client().Do(req)
		assert.NoError(t, err, "DELETE error")
		assert.Equal(t, r.StatusCode, 200, "DELETE status code")

		r2, err := server.Client().Get(server.URL + "/books/978-0134190440")
		assert.NoError(t, err, "GET error")
		assert.Equal(t, r2.StatusCode, 404, "GET status code")
	})
}

func ptr[T any](v T) *T {
	return &v
}

func unmarshalJSON[T any](t *testing.T, r *http.Response) T {
	t.Helper()
	defer r.Body.Close()

	var v T
	err := json.NewDecoder(r.Body).Decode(&v)
	assert.NoError(t, err, "unmarshalJSON error")

	return v
}

func newMockBookstore(t *testing.T, books []bookstore.Book) db.Bookstore {
	store, err := db.NewTestBookstore()
	assert.NoError(t, err, "NewTestBookstore error")

	for _, book := range books {
		err := store.AddBook(context.Background(), book)
		assert.NoError(t, err, "AddBook error")
	}

	return store
}

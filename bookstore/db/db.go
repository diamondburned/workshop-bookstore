package db

import (
	"context"
	"io"
	"path/filepath"

	"libdb.so/workshop-bookstore/bookstore"
)

// Bookstore represents a storage for books.
type Bookstore interface {
	io.Closer

	// Book retrieves a book by its ISBN.
	Book(ctx context.Context, isbn bookstore.ISBN) (bookstore.Book, error)
	// Books retrieves all books.
	Books(ctx context.Context) ([]bookstore.PartialBook, error)
	// AddBook adds a new book.
	AddBook(ctx context.Context, b bookstore.Book) error
	// UpdateBook updates an existing book.
	UpdateBook(ctx context.Context, isbn bookstore.ISBN, b UpdateBook) error
	// DeleteBook deletes a book by its ISBN.
	DeleteBook(ctx context.Context, isbn bookstore.ISBN) error
}

// UpdateBook represents a book update.
type UpdateBook struct {
	Title     *string                   `json:"title,omitempty"`
	Author    *string                   `json:"author,omitempty"`
	Price     *bookstore.Cents          `json:"price,omitempty"`
	Rating    **float64                 `json:"rating,omitempty"`
	Summary   **string                  `json:"summary,omitempty"`
	Language  **string                  `json:"language,omitempty"`
	Published **bookstore.PublishedDate `json:"published_date,omitempty"` // YYYY-MM-DD
}

// NewBookstore creates a new Bookstore.
func NewBookstore(path string) (Bookstore, error) {
	return newBadgerImpl(filepath.Join(path, "v1"))
}

// NewTestBookstore creates a new Bookstore for testing.
func NewTestBookstore() (Bookstore, error) {
	return newBadgerInMemoryImpl()
}

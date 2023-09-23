// Package bookstore describes the models for a bookstore.
package bookstore

import (
	"fmt"
	"regexp"
)

// Cents represents a price in USD cents.
type Cents int64

// String returns a string representation of the price. If a Price is used in
// fmt.Println, the String method will be called automatically.
func (p Cents) String() string {
	dollar := p / 100
	cents := p % 100
	return fmt.Sprintf("$%d.%02d", dollar, cents)
}

// ISBN represents a book's ISBN in ISBN-13 format.
type ISBN string

var isbnRe = regexp.MustCompile(`^\d{3}-\d+$`)

// Validate validates that the ISBN is valid. It returns an error if the ISBN is
// not.
func (isbn ISBN) Validate() error {
	if !isbnRe.MatchString(string(isbn)) {
		return fmt.Errorf("invalid ISBN: %s", isbn)
	}
	return nil
}

// PublishedDate represents a book's published date in YYYY-MM-DD format.
type PublishedDate string

// Parse parses the published date into its components.
func (p PublishedDate) Parse() (year, month, day int, err error) {
	_, err = fmt.Sscanf(string(p), "%d-%d-%d", &year, &month, &day)
	return
}

// PartialBook represents a partial book.
type PartialBook struct {
	ISBN   ISBN   `json:"isbn"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Price  Cents  `json:"price"`
}

func (b PartialBook) Validate() error {
	if err := b.ISBN.Validate(); err != nil {
		return err
	}
	if b.Title == "" {
		return fmt.Errorf("title is required")
	}
	if b.Author == "" {
		return fmt.Errorf("author is required")
	}
	if b.Price < 0 {
		return fmt.Errorf("price must be positive")
	}
	return nil
}

// Book represents a book.
type Book struct {
	PartialBook
	Rating    *float64       `json:"rating"`
	Summary   *string        `json:"summary"`
	Language  *string        `json:"language"`
	Published *PublishedDate `json:"published_date"` // YYYY-MM-DD
}

func (b Book) Validate() error {
	return b.PartialBook.Validate()
}

// ErrBookNotFound is returned when a book is not found.
var ErrBookNotFound = fmt.Errorf("book not found")

package db

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"libdb.so/workshop-bookstore/bookstore"
)

type badgerImpl struct {
	db *badger.DB
}

var _ Bookstore = (*badgerImpl)(nil)

func badgerOpts(path string) badger.Options {
	opts := badger.DefaultOptions(path)
	opts = opts.WithLoggingLevel(badger.WARNING)
	opts = opts.WithNumGoroutines(1)
	opts = opts.WithSyncWrites(true)
	return opts
}

func newBadgerImpl(path string) (*badgerImpl, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("creating directory: %w", err)
	}

	db, err := badger.Open(badgerOpts(path))
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	return &badgerImpl{db: db}, nil
}

func newBadgerInMemoryImpl() (*badgerImpl, error) {
	db, err := badger.Open(badgerOpts("").WithInMemory(true))
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	return &badgerImpl{db: db}, nil
}

func (db *badgerImpl) Close() error {
	return db.db.Close()
}

func (db *badgerImpl) Book(ctx context.Context, isbn bookstore.ISBN) (b bookstore.Book, err error) {
	err = db.db.View(func(tx *badger.Txn) error {
		return unmarshalFromTx(tx, keys("books", isbn), &b)
	})
	return b, badgerError(err)
}

func (db *badgerImpl) Books(ctx context.Context) (books []bookstore.PartialBook, err error) {
	books = []bookstore.PartialBook{}
	err = db.db.View(func(tx *badger.Txn) error {
		iopts := badger.DefaultIteratorOptions
		iopts.Prefix = keys("books")

		iter := tx.NewIterator(iopts)
		defer iter.Close()

		for iter.Rewind(); iter.Valid(); iter.Next() {
			var b bookstore.PartialBook
			if err := unmarshalItem(iter.Item(), &b); err != nil {
				return err
			}
			books = append(books, b)
		}

		return nil
	})
	return books, badgerError(err)
}

func (db *badgerImpl) AddBook(ctx context.Context, b bookstore.Book) error {
	err := db.db.Update(func(tx *badger.Txn) error {
		if _, err := tx.Get(keys("books", b.ISBN)); err == nil {
			return errors.New("book already exists")
		}
		return marshalIntoTx(tx, keys("books", b.ISBN), b)
	})
	return badgerError(err)
}

func (db *badgerImpl) UpdateBook(ctx context.Context, isbn bookstore.ISBN, update UpdateBook) error {
	err := db.db.Update(func(tx *badger.Txn) error {
		var old bookstore.Book
		if err := unmarshalFromTx(tx, keys("books", isbn), &old); err != nil {
			return err
		}

		setMaybe(&old.Title, update.Title)
		setMaybe(&old.Author, update.Author)
		setMaybe(&old.Price, update.Price)
		setMaybe(&old.Rating, update.Rating)
		setMaybe(&old.Summary, update.Summary)
		setMaybe(&old.Language, update.Language)
		setMaybe(&old.Published, update.Published)

		return marshalIntoTx(tx, keys("books", isbn), old)
	})
	return badgerError(err)
}

func (db *badgerImpl) DeleteBook(ctx context.Context, isbn bookstore.ISBN) error {
	err := db.db.Update(func(tx *badger.Txn) error {
		return tx.Delete(keys("books", isbn))
	})
	return badgerError(err)
}

func setMaybe[T any](dst *T, src *T) {
	if src != nil {
		*dst = *src
	}
}

func setMaybeNullable[T any](dst **T, src *T) {
	if src != nil {
		*dst = src
	}
}

func keys(v ...any) []byte {
	var b bytes.Buffer
	for i, v := range v {
		if i > 0 {
			b.WriteByte(0)
		}
		s, ok := v.(string)
		if ok {
			b.WriteString(s)
		} else {
			b.WriteString(fmt.Sprint(v))
		}
	}
	return b.Bytes()
}

func unmarshalItem(item *badger.Item, v any) error {
	return item.Value(func(data []byte) error {
		return json.Unmarshal(data, v)
	})
}

func unmarshalFromTx(tx *badger.Txn, k []byte, v any) error {
	item, err := tx.Get(k)
	if err != nil {
		return err
	}
	return unmarshalItem(item, v)
}

func marshalIntoTx(tx *badger.Txn, k []byte, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling: %w", err)
	}
	return tx.Set(k, b)
}

func badgerError(err error) error {
	if errors.Is(err, badger.ErrKeyNotFound) {
		return bookstore.ErrBookNotFound
	}
	return err
}

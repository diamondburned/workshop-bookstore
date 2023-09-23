package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"libdb.so/hrt"
	"libdb.so/workshop-bookstore/bookstore"
	"libdb.so/workshop-bookstore/bookstore/db"
)

// BookstoreHandler is the handler for the bookstore API.
type BookstoreHandler struct {
	// Embed ("inherit"-ish) chi.Router so we get http.Handler for free.
	chi.Router

	// store is the datastore for the bookstore. It is lower-case so the field
	// is private (unexported).
	store db.Bookstore
}

// NewBookstoreHandler creates a new BookstoreHandler.
func NewBookstoreHandler(store db.Bookstore) *BookstoreHandler {
	r := chi.NewRouter()
	h := &BookstoreHandler{
		Router: r,
		store:  store,
	}

	r.Use(hrt.Use(hrt.DefaultOpts))

	r.Route("/books", func(r chi.Router) {
		r.Get("/", hrt.Wrap(h.getBooks))
		r.Post("/", hrt.Wrap(h.addBook))
		r.Get("/{isbn}", hrt.Wrap(h.getBook))
		r.Patch("/{isbn}", hrt.Wrap(h.patchBook))
		r.Delete("/{isbn}", hrt.Wrap(h.deleteBook))
	})

	return h
}

func (h *BookstoreHandler) getBooks(ctx context.Context, _ hrt.None) ([]bookstore.PartialBook, error) {
	books, err := h.store.Books(ctx)
	if err != nil {
		return nil, apiError(err)
	}
	return books, nil
}

func (h *BookstoreHandler) getBook(ctx context.Context, _ hrt.None) (*bookstore.Book, error) {
	isbn := bookstore.ISBN(chi.URLParamFromCtx(ctx, "isbn"))
	if err := isbn.Validate(); err != nil {
		return nil, hrt.WrapHTTPError(http.StatusBadRequest, err)
	}

	book, err := h.store.Book(ctx, isbn)
	if err != nil {
		return nil, apiError(err)
	}

	return &book, nil
}

func (h *BookstoreHandler) addBook(ctx context.Context, book bookstore.Book) (hrt.None, error) {
	if err := h.store.AddBook(ctx, book); err != nil {
		return hrt.Empty, hrt.WrapHTTPError(http.StatusInternalServerError, err)
	}
	return hrt.Empty, nil
}

func (h *BookstoreHandler) patchBook(ctx context.Context, update db.UpdateBook) (*bookstore.Book, error) {
	isbn := bookstore.ISBN(chi.URLParamFromCtx(ctx, "isbn"))
	if err := isbn.Validate(); err != nil {
		return nil, hrt.WrapHTTPError(http.StatusBadRequest, err)
	}

	if err := h.store.UpdateBook(ctx, isbn, update); err != nil {
		return nil, apiError(err)
	}

	return h.getBook(ctx, hrt.None{})
}

func (h *BookstoreHandler) deleteBook(ctx context.Context, _ hrt.None) (hrt.None, error) {
	isbn := bookstore.ISBN(chi.URLParamFromCtx(ctx, "isbn"))
	if err := isbn.Validate(); err != nil {
		return hrt.Empty, hrt.WrapHTTPError(http.StatusBadRequest, err)
	}

	if err := h.store.DeleteBook(ctx, isbn); err != nil {
		return hrt.Empty, apiError(err)
	}

	return hrt.Empty, nil
}

func apiError(err error) error {
	status := http.StatusBadRequest
	if errors.Is(err, bookstore.ErrBookNotFound) {
		status = http.StatusNotFound
	}
	return hrt.WrapHTTPError(status, err)
}

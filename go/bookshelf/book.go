// Copyright 2015 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package bookshelf

import "context"

// Book holds metadata about a book.
type Book struct {
	ID            int64
	Title         string
	Author        string
	PublishedDate string
	ImageURL      string
	Description   string
	CreatedBy     string
	CreatedByID   string
}

// CreatedByDisplayName returns a string appropriate for displaying the name of
// the user who created this book object.
func (b *Book) CreatedByDisplayName() string {
	if b.CreatedByID == "anonymous" {
		return "Anonymous"
	}
	return b.CreatedBy
}

// SetCreatorAnonymous sets the CreatedByID field to the "anonymous" ID.
func (b *Book) SetCreatorAnonymous() {
	b.CreatedBy = ""
	b.CreatedByID = "anonymous"
}

// BookDatabase provides thread-safe access to a database of books.
type BookDatabase interface {
	// ListBooks returns a list of books, ordered by title.
	ListBooks(context.Context) ([]*Book, error)

	// ListBooksCreatedBy returns a list of books, ordered by title, filtered by
	// the user who created the book entry.
	ListBooksCreatedBy(ctx context.Context, userID string) ([]*Book, error)

	// GetBook retrieves a book by its ID.
	GetBook(ctx context.Context, id int64) (*Book, error)

	// AddBook saves a given book, assigning it a new ID.
	AddBook(ctx context.Context, b *Book) (id int64, err error)

	// DeleteBook removes a given book by its ID.
	DeleteBook(ctx context.Context, id int64) error

	// UpdateBook updates the entry for a given book.
	UpdateBook(ctx context.Context, b *Book) error

	// Close closes the database, freeing up any available resources.
	// TODO(cbro): Close() should return an error.
	Close(ctx context.Context)
}

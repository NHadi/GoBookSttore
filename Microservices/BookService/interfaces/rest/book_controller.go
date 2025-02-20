package rest

import (
	"encoding/json"
	"net/http"

	"BookStore/application/commands"
	"BookStore/application/queries"
	"BookStore/application/services"
	"BookStore/domain/entities"

	"github.com/gorilla/mux"
)

// BookController is a REST controller for book-related actions.
type BookController struct {
	service *services.BookApplicationService
}

// NewBookController creates a new BookController instance.
func NewBookController(service *services.BookApplicationService) *BookController {
	return &BookController{service: service}
}

// CreateBook godoc
// @Summary Create a new book
// @Description Create a new book with the provided details
// @Tags books
// @Accept json
// @Produce json
// @Param book body commands.AddBookCommand true "Add book command"
// @Success 201 {object} entities.Book
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /books [post]
func (c *BookController) CreateBook(w http.ResponseWriter, r *http.Request) {
	var cmd commands.AddBookCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	book, err := c.service.AddBook(cmd)
	if err != nil {
		http.Error(w, "Error adding book", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (c *BookController) SearchBooks(w http.ResponseWriter, r *http.Request) {
	// Get the query parameter
	query := r.URL.Query().Get("query")

	// Call the service to search for books
	books, err := c.service.SearchBooks(query)
	if err != nil {
		http.Error(w, "Error searching for books", http.StatusInternalServerError)
		return
	}

	// Prepare the response with pointers
	bookPointers := make([]*entities.Book, len(books))
	for i := range books {
		bookPointers[i] = &books[i] // Convert each book to a pointer
	}

	response := BooksResponse{
		Count: len(bookPointers),
		Books: bookPointers, // Now this is a slice of pointers
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetBooks godoc
// @Summary List all books
// @Description Get a list of all books
// @Tags books
// @Produce json
// @Success 200 {object} BooksResponse
// @Failure 500 {object} ErrorResponse
// @Router /books [get]
func (c *BookController) GetBooks(w http.ResponseWriter, r *http.Request) {
	query := queries.GetBooksQuery{}
	books := c.service.ListBooks(query)

	response := BooksResponse{
		Count: len(books),
		Books: books,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateBook godoc
// @Summary Update a book
// @Description Update details of a specific book by its ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path string true "Book ID"
// @Param book body entities.Book true "Updated book info"
// @Success 200 {object} entities.Book
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /books/{id} [put]
func (c *BookController) UpdateBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	book, err := c.service.FindByID(id)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(book); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := c.service.Update(book); err != nil {
		http.Error(w, "Error updating book", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(book)
}

// DeleteBook godoc
// @Summary Delete a book
// @Description Delete a specific book by its ID
// @Tags books
// @Param id path string true "Book ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Router /books/{id} [delete]
func (c *BookController) DeleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if err := c.service.Delete(id); err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// BorrowBook godoc
// @Summary Borrow a book
// @Description Borrow a book for a user
// @Tags books
// @Produce json
// @Param userId path string true "User ID"
// @Param bookId path string true "Book ID"
// @Success 200 {string} string "message"
// @Failure 403 {string} string "error"
// @Failure 500 {string} string "error"
// @Router /books/borrow/{userId}/{bookId} [post]
func (c *BookController) BorrowBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]
	bookId := vars["bookId"]

	message, err := c.service.BorrowBook(userId, bookId)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(message))
}

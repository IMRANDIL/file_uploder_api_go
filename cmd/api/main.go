package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/cors"
)

const port = 8080

func main() {
	mux := http.NewServeMux()

	// Register a handler function for the root path ("/")
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from Go backend")
	})

	// Create a new CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Wrap the existing mux with the CORS middleware
	handler := c.Handler(mux)

	// Register the upload handler
	mux.HandleFunc("/api/v1/upload", uploadHandler)

	log.Println("app started on port", port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
	if err != nil {
		log.Fatal(err)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// fmt.Println(r.Header)
	// Check if the content type is multipart/form-data
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		http.Error(w, "Unsupported media type", http.StatusUnsupportedMediaType)
		return
	}

	// Set a limit of 5MB for the uploaded file
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	// Parse the form data
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		http.Error(w, "Max file size is 5mb", http.StatusBadRequest)
		return
	}

	// Get the file from the request
	file, header, err := r.FormFile("file")

	if err != nil {
		http.Error(w, "File not found", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check that the file is a PNG, JPG, or JPEG
	ext := filepath.Ext(header.Filename)
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		http.Error(w, "Only PNG, JPG and JPEG format allowed", http.StatusBadRequest)
		return
	}

	// Create the directory if it doesn't exist
	err = os.MkdirAll("./images", 0755)
	if err != nil {
		http.Error(w, "Error creating images directory", http.StatusInternalServerError)
		return
	}

	// Create the file in the images folder
	path := "./images/" + header.Filename
	out, err := os.Create(path)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Write the file to disk using a stream
	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "File uploaded successfully")
}

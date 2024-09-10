package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	port = ":8080"
)

func main() {
	http.Handle("GET /", http.FileServer(http.Dir("./static")))

	http.HandleFunc("POST /upload", handleFileUpload)

	http.HandleFunc("POST /message", handleMessage)

	// Define the HTTPS server configuration
	server := &http.Server{
		Addr: port,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}

	// Start the HTTPS server
	fmt.Printf("Starting server on %s\n", port)
	if err := server.ListenAndServeTLS("secrets/server.crt", "secrets/server.key"); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

// handleFileUpload processes file uploads
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Start handling request...")
	start := time.Now().UnixMilli()

	// Parse the form to retrieve the file
	err := r.ParseMultipartForm(1024 * 1024 * 1024) // 1 GB limit
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, fileHandler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a file on the server to save the uploaded file
	saveFileName := fmt.Sprintf("%d-%s", time.Now().UnixMilli(), fileHandler.Filename)
	out, err := os.Create(fmt.Sprintf("./uploaded/%s", saveFileName))
	if err != nil {
		http.Error(w, "Unable to create file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Copy the file content to the new file
	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/success.html", http.StatusSeeOther)

	fmt.Printf("Took %d ms\n\n", time.Now().UnixMilli()-start)
}

type Message struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	TimeSent
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handling short message")

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Failed to read message body", http.StatusBadRequest)
		return
	}

	var message Message
	err = json.Unmarshal(body, &message)
	if err != nil {
		http.Error(w, "Invalid message body", http.StatusBadRequest)
		return
	}

	messageLogFile, err := os.OpenFile("uploaded/messages.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		fmt.Println(fmt.Errorf("Failed to open message log file: %w\n", err))
		return
	}
	defer messageLogFile.Close()

	messageLogFile.WriteString(fmt.Sprintf("%s:%s\n", message.Title, message.Body))

	fmt.Printf("Got message: %+v\n", message)

	fmt.Fprintln(w, "Message sent")
}

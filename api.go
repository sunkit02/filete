package main

import (
	"crypto/tls"
	"encoding/json"
	"floader/data"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ServerConfigs struct {
	Port     uint16
	CertFile string
	KeyFile  string
	AssetDir string
}

func StartServer(configs ServerConfigs) {
	http.Handle("GET /", http.FileServer(http.Dir(configs.AssetDir)))

	http.HandleFunc("POST /upload", handleFileUpload)

	http.HandleFunc("POST /message", handleMessage)

	// Define the HTTPS server configuration
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", configs.Port),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}

	// Start the HTTPS server
	Info.Printf("Start listening on port %d with TLS\n", configs.Port)
	if err := server.ListenAndServeTLS(configs.CertFile, configs.KeyFile); err != nil {
		Error.Fatalf("Error starting server: %v\n", err)
	}
}

// handleFileUpload processes file uploads
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	id := uuid.New()

	Info.Println(withId(id, reqLogStr(r, false)))
	Debug.Println(withId(id, reqLogStr(r, true)))

	uploadedFile, uploadedFileHandler, err := r.FormFile("file")
	if err != nil {
		msg := withId(id, "Unable to read file from request")
		Error.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	defer uploadedFile.Close()

	// Create a file on the server to save the uploaded file
	saveFilePath := fmt.Sprintf("./uploaded/%d-%s", time.Now().UnixMilli(), uploadedFileHandler.Filename)
	saveFile, err := os.Create(saveFilePath)
	if err != nil {
		Error.Printf(withId(id, "Unable to save file with name '%s'", saveFilePath))
		http.Error(w, withId(id, "Error saving file"), http.StatusInternalServerError)
		return
	}
	defer saveFile.Close()

	// Copy the file content to the new file
	bytesWritten, err := io.Copy(saveFile, uploadedFile)
	if err != nil {
		Error.Printf(withId(id, "Unable to copy file with name '%s'", saveFilePath))
		http.Error(w, withId(id, "Error saving file"), http.StatusInternalServerError)
		return
	}

	// Remove savefile if incomplete/corrupted
	if bytesWritten != uploadedFileHandler.Size {
		err = os.Remove(saveFilePath)

		Error.Printf(withId(id,
			"Saved file incomplete/corrupted. Uploaded size: %d, written: %d",
			bytesWritten, uploadedFileHandler.Size))

		http.Error(w, withId(id, "Error saving file"), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/success.html", http.StatusSeeOther)
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	id := uuid.New()

	Info.Println(withId(id, reqLogStr(r, false)))
	Debug.Println(withId(id, reqLogStr(r, true)))

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		Error.Println(withId(id, "Failed to read request body"))
		http.Error(w, withId(id, "Invalid message body"), http.StatusBadRequest)
		return
	}

	var message data.Message
	err = json.Unmarshal(body, &message)
	if err != nil {
		Error.Println(withId(id, "Failed to parse Message from request body"))
		http.Error(w, withId(id, "Invalid message body"), http.StatusBadRequest)
		return
	}

	messageLogFile, err := os.OpenFile("uploaded/messages.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Error.Println(withId(id, "Failed to open message log file: %w\n", err))
		http.Error(w, withId(id, "Internal Server Error"), 500)
		return
	}
	defer messageLogFile.Close()

	messageLogFile.WriteString(fmt.Sprintf("%v#|#%v#|#%v\n", message.Title, message.Body, message.TimeSent.UTC().UTC().Format(time.RFC3339)))

	fmt.Printf("Got message: %+v\n", message)

	fmt.Fprintln(w, "Message sent")
}

func reqLogStr(r *http.Request, headers bool) string {
	var logStr string
	var sep = ""
	if len(r.URL.RawQuery) > 0 {
		sep = "?"
	}
	path := fmt.Sprintf("%s%s%s", r.URL.Path, sep, r.URL.RawQuery)

	if headers {
		var headerSb strings.Builder
		headerSb.WriteString("Headers {")
		for k, vals := range r.Header {
			headerSb.WriteString(k)
			headerSb.WriteString("=[")
			for i, v := range vals {
				headerSb.WriteString(v)
				if i < len(vals)-1 {
					headerSb.WriteString(", ")
				}
			}
			headerSb.WriteString("], ")
		}
		headerSb.WriteRune('}')

		logStr = fmt.Sprintf("%s %s %s", r.Method, path, headerSb.String())
	} else {
		logStr = fmt.Sprintf("%s %v", r.Method, path)
	}
	return logStr
}

func withId(id uuid.UUID, format string, a ...any) string {
	return fmt.Sprintf("reqId(%s) "+format, id, a)
}

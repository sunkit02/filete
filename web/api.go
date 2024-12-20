package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sunkit02/filete/data"
	"github.com/sunkit02/filete/logging"
	"github.com/sunkit02/filete/services"
	"github.com/sunkit02/filete/web/middleware"
	"github.com/sunkit02/filete/web/utils"
)

func ApiRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /upload", handleFileUpload)
	mux.HandleFunc("POST /message", handlePostMessage)
	mux.HandleFunc("GET /shared-dir", handleGetSharedDir)
	mux.HandleFunc("GET /download", handleFileDownload)

	return mux
}

const MaxFileSizeStoredInMemory = 32 << 20 // 32 MB

// handleFileUpload processes file uploads
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	id := middleware.ExtractRequestId(r)

	err := r.ParseMultipartForm(MaxFileSizeStoredInMemory)
	if err != nil {
		logging.Error.Println(utils.WithId(id, "Failed to parse multipart file"))
		http.Error(w, utils.WithId(id, "Failed to parse multipart file upload"), http.StatusBadRequest)
		return
	}

	if len(r.MultipartForm.File) == 0 {
		logging.Error.Println(utils.WithId(id, "No multipart files found"))
		http.Error(w, utils.WithId(id, "No files found"), http.StatusBadRequest)
		return
	}

	for k, fileheaders := range r.MultipartForm.File {
		logging.Debug.Println("Processing multipart fileheaders with key:", k)

		for _, fileheader := range fileheaders {
			uploadedFile, err := fileheader.Open()
			if err != nil {
				logging.Error.Println(utils.WithId(id, "Failed to parse multipart file upload: %v", err))
				http.Error(w, utils.WithId(id, "Failed to parse multipart file upload"), http.StatusBadRequest)
				return
			}

			logging.Trace.Println("Processing uploaded file:", fileheader.Filename)

			// Create a file on the server to save the uploaded file
			saveFilePath := fmt.Sprintf("./uploaded/%d-%s", time.Now().UnixMilli(), fileheader.Filename)
			saveFile, err := os.Create(saveFilePath)
			if err != nil {
				logging.Error.Printf(utils.WithId(id, "Unable to save file with name '%s'", saveFilePath))
				http.Error(w, utils.WithId(id, "Error saving file"), http.StatusInternalServerError)
				return
			}
			defer saveFile.Close()

			// Copy the file content to the new file
			bytesWritten, err := io.Copy(saveFile, uploadedFile)
			if err != nil {
				logging.Error.Printf(utils.WithId(id, "Unable to copy file with name '%s'", saveFilePath))
				http.Error(w, utils.WithId(id, "Error saving file"), http.StatusInternalServerError)
				return
			}

			// Remove savefile if incomplete/corrupted
			if bytesWritten != fileheader.Size {
				err = os.Remove(saveFilePath)

				logging.Error.Printf(utils.WithId(id,
					"Saved file incomplete/corrupted. Uploaded size: %d, written: %d",
					bytesWritten, fileheader.Size))

				http.Error(w, utils.WithId(id, "Error saving file"), http.StatusInternalServerError)
				return
			}

		}
	}

	fmt.Fprint(w, "File(s) uploaded successfully")
}

func handlePostMessage(w http.ResponseWriter, r *http.Request) {
	id := middleware.ExtractRequestId(r)

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		logging.Error.Println(utils.WithId(id, "Failed to read request body"))
		http.Error(w, utils.WithId(id, "Invalid message body"), http.StatusBadRequest)
		return
	}

	var message data.Message
	err = json.Unmarshal(body, &message)
	if err != nil {
		logging.Error.Println(utils.WithId(id, "Failed to parse Message from request body"))
		http.Error(w, utils.WithId(id, "Invalid message body"), http.StatusBadRequest)
		return
	}

	err = messageRepo.Add(message)
	if err != nil {
		logging.Error.Println(utils.WithId(id, "Failed to write new message to file"), err)
		http.Error(w, utils.WithId(id, "Failed to send new message"), http.StatusInternalServerError)
	}

	fmt.Printf("Got message: %+v\n", message)

	fmt.Fprint(w, "Message sent")
}

const DefaultReadDepth = 1

// Getting with query parameter `path` empty gets the sharedRootDirectories in an array.
// If `path` is empty, query parameter `root-dir-hash` is ignored
func handleGetSharedDir(w http.ResponseWriter, r *http.Request) {
	id := middleware.ExtractRequestId(r)

	path := r.URL.Query().Get("path")

	var responseBody []byte
	if path == "" {
		sharedDirs, err := services.ReadRootDirs(DefaultReadDepth)
		if err != nil {
			logging.Error.Println(utils.WithId(id, err.Error()))
			http.Error(w, utils.WithId(id, err.Error()), http.StatusInternalServerError)
			return
		}
		responseBody, err = json.Marshal(sharedDirs)
		if err != nil {
			logging.Error.Println(utils.WithId(id, err.Error()))
			http.Error(w, utils.WithId(id, err.Error()), http.StatusInternalServerError)
			return
		}
	} else {
		rootDirHash := r.URL.Query().Get("root-dir-hash")
		sharedDir, err := services.ReadDir(path, rootDirHash, DefaultReadDepth)
		if err != nil {
			logging.Error.Println(utils.WithId(id, err.Error()))
			http.Error(w, utils.WithId(id, err.Error()), http.StatusInternalServerError)
			return
		}
		responseBody, err = json.Marshal(sharedDir)
		if err != nil {
			logging.Error.Println(utils.WithId(id, err.Error()))
			http.Error(w, utils.WithId(id, err.Error()), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")

	written, err := w.Write(responseBody)
	if err != nil {
		logging.Error.Println(utils.WithId(id, err.Error()))
		return
	} else if written != len(responseBody) {
		logging.Error.Println(utils.WithId(id,
			"Failed to write full response body. All: %d, written:%d",
			len(responseBody), written))
	}
}

func handleFileDownload(w http.ResponseWriter, r *http.Request) {
	id := middleware.ExtractRequestId(r)

	path := r.URL.Query().Get("path")
	rootDirHash := r.URL.Query().Get("root-dir-hash")

	if rootDirHash == "" {
		http.Error(w, utils.WithId(id, "Invalid path or rootDirHash"), http.StatusBadRequest)
		return
	}

	file, fileName, isDir, err := services.GetFileForDownload(path, rootDirHash)
	if err != nil {
		logging.Error.Println(utils.WithId(id, err.Error()))
		http.Error(w, utils.WithId(id, "Internal error"), http.StatusInternalServerError)
		return
	}

	var contentType string
	if isDir {
		contentType = "application/zip"
	} else {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; file-name="+fileName)

	// TODO: Check for complete file transfer
	_, err = io.Copy(w, file)
	if err != nil {
		logging.Error.Println(utils.WithId(id, err.Error()))
		return
	}
}

// TODO: Get rid of this security vulnerability
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

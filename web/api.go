package web

import (
	"crypto/tls"
	"encoding/json"
	"floader/data"
	"floader/logging"
	"floader/services"
	"floader/web/middleware"
	"fmt"
	"io"
	"math/rand"
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

	// Path to directory holding static assets
	AssetDir string

	// Path to directory that saves all uploaded content
	// NOTE: Must be pointing to a directory or empty
	UploadDir string

	// Path to directories to be shared
	ShareDirs []string

	// Key required to be entered by client to authenticate. The server will
	// generate a random one if left empty.
	SessionKey string
}

var (
	sessionKey  string
	messageRepo data.Repository[data.MessageId, data.Message]
)

func StartServer(configs ServerConfigs) {
	r := data.NewFileMessageRepo(configs.UploadDir + "/messages.dat")
	messageRepo = &r

	services.InitDownloadService(services.DownloadServiceConfig{
		SharedDirectories: configs.ShareDirs,
	})

	if configs.SessionKey == "" {
		configs.SessionKey = generateSessionKey(8)
	}
	sessionKey = configs.SessionKey

	assetMux := http.NewServeMux()
	assetMux.Handle("GET /", http.FileServer(http.Dir(configs.AssetDir)))

	apiMux := http.NewServeMux()
	apiMux.HandleFunc("POST /upload", handleFileUpload)
	apiMux.HandleFunc("POST /message", handlePostMessage)
	apiMux.HandleFunc("GET /shared-dir", handleGetSharedDir)

	composedMux := http.NewServeMux()
	composedMux.Handle("/", assetMux)
	composedMux.Handle("/api/", middleware.SessionKeyAuthMiddleware(
		sessionKey,
		http.StripPrefix("/api", apiMux),
	))

	topLevelMux := http.NewServeMux()
	topLevelMux.Handle("/",
		middleware.RequestIdMiddleware(
			middleware.RequestLoggingMiddleware(composedMux)))

	// Define the HTTPS server configuration
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", configs.Port),
		Handler: topLevelMux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}

	// Start the HTTPS server
	logging.Info.Printf("Start listening on port %d with TLS\n", configs.Port)
	logging.Info.Println("Session key:", sessionKey)
	if err := server.ListenAndServeTLS(configs.CertFile, configs.KeyFile); err != nil {
		logging.Error.Fatalf("Error starting server: %v\n", err)
	}
}

const MaxFileSizeStoredInMemory = 32 << 20 // 32 MB

// handleFileUpload processes file uploads
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	id := middleware.ExtractRequestId(r)

	err := r.ParseMultipartForm(MaxFileSizeStoredInMemory)
	if err != nil {
		logging.Error.Println(withId(id, "Failed to parse multipart file"))
		http.Error(w, withId(id, "Failed to parse multipart file upload"), http.StatusBadRequest)
		return
	}

	if len(r.MultipartForm.File) == 0 {
		logging.Error.Println(withId(id, "No multipart files found"))
		http.Error(w, withId(id, "No files found"), http.StatusBadRequest)
		return
	}

	for k, fileheaders := range r.MultipartForm.File {
		logging.Debug.Println("Processing multipart fileheaders with key:", k)

		for _, fileheader := range fileheaders {
			uploadedFile, err := fileheader.Open()
			if err != nil {
				logging.Error.Println(withId(id, "Failed to parse multipart file upload: %v", err))
				http.Error(w, withId(id, "Failed to parse multipart file upload"), http.StatusBadRequest)
				return
			}

			logging.Trace.Println("Processing uploaded file:", fileheader.Filename)

			// Create a file on the server to save the uploaded file
			saveFilePath := fmt.Sprintf("./uploaded/%d-%s", time.Now().UnixMilli(), fileheader.Filename)
			saveFile, err := os.Create(saveFilePath)
			if err != nil {
				logging.Error.Printf(withId(id, "Unable to save file with name '%s'", saveFilePath))
				http.Error(w, withId(id, "Error saving file"), http.StatusInternalServerError)
				return
			}
			defer saveFile.Close()

			// Copy the file content to the new file
			bytesWritten, err := io.Copy(saveFile, uploadedFile)
			if err != nil {
				logging.Error.Printf(withId(id, "Unable to copy file with name '%s'", saveFilePath))
				http.Error(w, withId(id, "Error saving file"), http.StatusInternalServerError)
				return
			}

			// Remove savefile if incomplete/corrupted
			if bytesWritten != fileheader.Size {
				err = os.Remove(saveFilePath)

				logging.Error.Printf(withId(id,
					"Saved file incomplete/corrupted. Uploaded size: %d, written: %d",
					bytesWritten, fileheader.Size))

				http.Error(w, withId(id, "Error saving file"), http.StatusInternalServerError)
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
		logging.Error.Println(withId(id, "Failed to read request body"))
		http.Error(w, withId(id, "Invalid message body"), http.StatusBadRequest)
		return
	}

	var message data.Message
	err = json.Unmarshal(body, &message)
	if err != nil {
		logging.Error.Println(withId(id, "Failed to parse Message from request body"))
		http.Error(w, withId(id, "Invalid message body"), http.StatusBadRequest)
		return
	}

	err = messageRepo.Add(message)
	if err != nil {
		logging.Error.Println(withId(id, "Failed to write new message to file"), err)
		http.Error(w, withId(id, "Failed to send new message"), http.StatusInternalServerError)
	}

	fmt.Printf("Got message: %+v\n", message)

	fmt.Fprint(w, "Message sent")
}

const DefaultReadDepth = 2

// Getting with query parameter `path` empty gets the sharedRootDirectories in an array.
// If `path` is empty, query parameter `root-dir-hash` is ignored
func handleGetSharedDir(w http.ResponseWriter, r *http.Request) {
	id := middleware.ExtractRequestId(r)

	path := r.URL.Query().Get("path")

	var responseBody []byte
	if path == "" {
		sharedDirs, err := services.ReadRootDirs(DefaultReadDepth)
		if err != nil {
			logging.Error.Println(withId(id, err.Error()))
			http.Error(w, withId(id, err.Error()), http.StatusInternalServerError)
			return
		}
		responseBody, err = json.Marshal(sharedDirs)
		if err != nil {
			logging.Error.Println(withId(id, err.Error()))
			http.Error(w, withId(id, err.Error()), http.StatusInternalServerError)
			return
		}
	} else {
		rootDirHash := r.URL.Query().Get("root-dir-hash")
		sharedDir, err := services.ReadDir(path, rootDirHash, DefaultReadDepth)
		if err != nil {
			logging.Error.Println(withId(id, err.Error()))
			http.Error(w, withId(id, err.Error()), http.StatusInternalServerError)
			return
		}
		responseBody, err = json.Marshal(sharedDir)
		if err != nil {
			logging.Error.Println(withId(id, err.Error()))
			http.Error(w, withId(id, err.Error()), http.StatusInternalServerError)
			return
		}
	}

	written, err := w.Write(responseBody)
	if err != nil {
		logging.Error.Println(withId(id, err.Error()))
		return
	} else if written != len(responseBody) {
		logging.Error.Println(withId(id,
			"Failed to write full response body. All: %d, written:%d",
			len(responseBody), written))
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

func withId(id uuid.UUID, format string, a ...any) string {
	args := make([]any, 0, len(a)+1)
	args = append(args, id)
	args = append(args, a...)

	return fmt.Sprintf("reqId(%s) "+format, args...)
}

var alphanumerics = [...]rune{
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I',
	'J', 'K', 'L', 'M', 'N', 'O', 'P', 'q', 'r',
	's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '1',
	'2', '3', '4', '5', '6', '7', '8', '9', '0',
}

func generateSessionKey(length uint) string {
	var key strings.Builder
	for range length {
		key.WriteRune(alphanumerics[rand.Intn(len(alphanumerics))])
	}

	return key.String()
}

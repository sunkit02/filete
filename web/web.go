package web

import (
	"crypto/tls"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/sunkit02/filete/data"
	"github.com/sunkit02/filete/logging"
	"github.com/sunkit02/filete/services"
	"github.com/sunkit02/filete/utils"
	"github.com/sunkit02/filete/web/middleware"
)

type ServerConfigs struct {
	Port     uint16
	CertFile string
	KeyFile  string

	// Path to directory holding static assets
	Assets fs.FS

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

	// Ensure that the session key is not empty
	if configs.SessionKey == "" {
		configs.SessionKey = utils.GenerateRandomString(8)
	}
	sessionKey = configs.SessionKey

	// Init services
	services.InitDownloadService(services.DownloadServiceConfig{
		SharedDirectories: configs.ShareDirs,
	})

	services.InitAuthService(services.AuthServiceConfig{
		SessionKey: configs.SessionKey,
	})

	// Initialize routes
	cookieAuthMiddleware := middleware.CookieAuthMiddleware

	composedMux := http.NewServeMux()
	composedMux.Handle("/", StaticAssetsRoute(configs))
	composedMux.Handle("/auth/", http.StripPrefix("/auth", AuthRoutes()))
	composedMux.Handle("/api/", cookieAuthMiddleware(
		http.StripPrefix("/api", ApiRoutes()),
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

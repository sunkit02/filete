package main

import (
	"embed"
	"io/fs"
	"os"

	"github.com/sunkit02/filete/logging"
	"github.com/sunkit02/filete/web"
)

//go:embed static/*
var EmbeddedAssets embed.FS

func init() {
	logging.InitializeLoggers(os.Stdout)
}

func main() {
	args := os.Args[1:]

	staticRoot, err := fs.Sub(EmbeddedAssets, "static")
	if err != nil {
		logging.Error.Fatal(err)
	}

	serverConfigs := web.ServerConfigs{
		Port:      8080,
		CertFile:  "./secrets/server.crt",
		KeyFile:   "./secrets/server.key",
		Assets:    staticRoot,
		ShareDirs: args,
		// ShareDirs:  []string{"/home/sunkit/src"},
		UploadDir:  "./uploaded",
		SessionKey: "123",
	}

	web.StartServer(serverConfigs)
}

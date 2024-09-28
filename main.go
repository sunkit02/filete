package main

import (
	"floader/logging"
	"os"
)

func init() {
	logging.InitializeLoggers(os.Stdout)
}

func main() {
	serverConfigs := ServerConfigs{
		Port:      8080,
		CertFile:  "./secrets/server.crt",
		KeyFile:   "./secrets/server.key",
		AssetDir:  "./static",
		UploadDir: "./uploaded",
	}

	StartServer(serverConfigs)
}

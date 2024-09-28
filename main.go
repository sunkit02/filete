package main

import (
	"floader/logging"
	"floader/web"
	"os"
)

func init() {
	logging.InitializeLoggers(os.Stdout)
}

func main() {
	serverConfigs := web.ServerConfigs{
		Port:      8080,
		CertFile:  "./secrets/server.crt",
		KeyFile:   "./secrets/server.key",
		AssetDir:  "./static",
		UploadDir: "./uploaded",
	}

	web.StartServer(serverConfigs)
}

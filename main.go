package main

import (
	"floader/data"
	"floader/logging"
	"os"
)

func init() {
	logging.InitializeLoggers(os.Stdout)
}

func main() {
	serverConfigs := ServerConfigs{
		Port:     8080,
		CertFile: "./secrets/server.crt",
		KeyFile:  "./secrets/server.key",
		AssetDir: "./static",
	}

	StartServer(serverConfigs)
}

func useRepo[K any, T any](r data.Repository[K, T]) {

}

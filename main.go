package main

import (
	"filete/logging"
	"filete/web"
	"os"
)

func init() {
	logging.InitializeLoggers(os.Stdout)
}

func main() {
	args := os.Args[1:]

	serverConfigs := web.ServerConfigs{
		Port:      8080,
		CertFile:  "./secrets/server.crt",
		KeyFile:   "./secrets/server.key",
		AssetDir:  "./static",
		ShareDirs: args,
		// ShareDirs:  []string{"/home/sunkit/src"},
		UploadDir:  "./uploaded",
		SessionKey: "123",
	}

	web.StartServer(serverConfigs)
}

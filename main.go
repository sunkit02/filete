package main

import "os"

func init() {
	initializeLoggers(os.Stdout)
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

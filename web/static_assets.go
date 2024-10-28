package web

import "net/http"

func StaticAssetsRoute(configs ServerConfigs) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.FS(configs.Assets)))
	return mux
}

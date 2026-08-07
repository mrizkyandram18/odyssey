package api

import (
	"log"
	"net/http"
	"os"
	"strings"

	"odyssey/pkg/server"
)

var handler http.Handler
var initErr error
var initialized bool

func Handler(w http.ResponseWriter, r *http.Request) {
	if !initialized {
		srvHandler, err := server.BuildHandler()
		if err != nil {
			initErr = err
			// Log all available environment variable keys for debugging
			var keys []string
			for _, e := range os.Environ() {
				pair := strings.SplitN(e, "=", 2)
				keys = append(keys, pair[0])
			}
			log.Printf("Server build error: %v. Available env keys: %v", err, keys)
		} else {
			handler = srvHandler.Handler
		}
		initialized = true
	}

	if initErr != nil {
		http.Error(w, "Server not initialized: "+initErr.Error(), http.StatusInternalServerError)
		return
	}
	handler.ServeHTTP(w, r)
}


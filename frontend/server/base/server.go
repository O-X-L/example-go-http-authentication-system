package base

import (
	"encoding/json"
	server "example_fe"
	"example_fe/base/db"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func Server(store *db.DataStore) {
	mux := http.NewServeMux()
	validator := validator.New()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	clientBuildFS, err := fs.Sub(server.EmbeddedStaticFiles, "files/build")
	if err != nil {
		log.Fatal("Could not find embedded build folder.", err)
	}
	clientBuildHTTPFS := http.FS(clientBuildFS)

	spa := &spaHandler{
		fileSystem: clientBuildHTTPFS,
		store:      store,
		validator:  validator,
	}
	wrappedHandler := guestMiddleware(
		store,
		validator,
		spa,
	)
	mux.Handle("/", wrappedHandler)

	port := ":8000"
	log.Printf("Server starting on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal(err)
	}
}

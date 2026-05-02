package base

import (
	"net/http"
	"strings"

	"example_fe/base/db"

	"github.com/go-playground/validator/v10"
)

var BUILD_FILE_PREFIXES = []string{
	"_app/",
	"favicon.png",
	"a/",
}

// spaHandler serves the embedded Svelte SPA
type spaHandler struct {
	fileSystem http.FileSystem
	store      *db.DataStore
	validator  *validator.Validate
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	// if not part of build-files => redirect to base-path
	if path == "" {
		path = "index.html"

	} else {
		isBuildPath := false
		for _, prefix := range BUILD_FILE_PREFIXES {
			if strings.HasPrefix(path, prefix) {
				isBuildPath = true
				break
			}
		}
		if !isBuildPath {
			http.Redirect(w, r, "/", 302)
			return
		}
	}

	// if file does not exist => 404
	f, err := h.fileSystem.Open(path)
	if err != nil {
		f, err = h.fileSystem.Open(path + ".html")
		if err == nil {
			r.URL.Path = "/" + path + ".html"

		} else {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
	}
	f.Close()

	w.Header().Add("Cache-Control", "max-age=600, must-revalidate") // 10m

	// serve the static file
	http.FileServer(h.fileSystem).ServeHTTP(w, r)
}

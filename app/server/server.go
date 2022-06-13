package server

import (
	"actionboard/app/data"
	"actionboard/web"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"
)

// Start launched the web server
func Start(data *data.Data, config data.Config) {

	var fSys fs.FS
	var err error

	_, doNotEmbedFiles := os.LookupEnv("DO_NOT_EMBED_FILES")
	if doNotEmbedFiles {
		fSys = os.DirFS("web/static")
	} else {
		fSys, err = fs.Sub(web.StaticFiles, "static")
		if err != nil {
			panic(err)
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", alwaysSecure(http.StripPrefix("/static/", http.FileServer(http.FS(fSys)))))
	mux.Handle("/", alwaysSecure(handleIndex(data, config)))
	mux.Handle("/pause", alwaysSecure(handlePause(data, config)))

	port := config.Port

	addr := fmt.Sprintf(":%s", port)
	server := http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	log.Println("main: running simple server on port", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("main: couldn't start simple server: %v\n", err)
	}
}

// alwaysSecure will redirect http requests to be https
func alwaysSecure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Forwarded-Proto") == "http" {
			http.Redirect(w, r, "https://" + r.Host + r.RequestURI, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

//fileSystem will server the html files from either embedded in the binary or separate
func fileSystem() fs.FS {

	var fsys fs.FS

	_, doNotEmbedFiles := os.LookupEnv("DO_NOT_EMBED_FILES")
	if doNotEmbedFiles {
		fsys = os.DirFS("web")
	} else {
		fsys = web.HtmlFiles
	}

	return fsys
}

func htmlFunctions() map[string]interface{} {
	return map[string]interface{}{"makeArray": makeArray, "add": add, "multiply": multiply}
}

func makeArray(args ...interface{}) []interface{} {
	return args
}

func add(a int, b int) int {
	return a + b
}

func multiply(a int, b int) int {
	return a * b
}

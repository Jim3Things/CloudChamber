// This module containes the routines etc to implement the frontend handlers for the files
// part of the API
//

//package frontend
package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
) //Comment from Manpreet. 

func filesAddRoutes(routeBase *mux.Router) {

	routeFiles := routeBase.PathPrefix("/static").Subrouter()

	routeBase.HandleFunc("", handlerFilesRoot).Methods("GET")
	routeBase.HandleFunc("/", handlerFilesRoot).Methods("GET")

	routeFiles.HandleFunc("", handlerFilesRoot).Methods("GET")
	routeFiles.HandleFunc("/", handlerFilesRoot).Methods("GET")
}

func handlerFilesRoot(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Files (Root)")
	//	serveFileTemplate(w, r)

	return
}

func serveFileErrorPage(code int, dir string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		// Call duplicated by src/net/http/fs.go#L254
		w.WriteHeader(code)

		path := filepath.Join(dir, r.URL.Path)
		log.Println(path)

		http.ServeFile(w, r, path)
	}

	return http.HandlerFunc(fn)
}

func serveFileStatic(path string) http.Handler {
	fp := filepath.Join(rootFilePath, path)

	return http.FileServer(http.Dir(fp))
}

func serveFileTemplate(w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join(rootFilePath, "templates", "layout.html")
	fp := filepath.Join(rootFilePath, "templates", filepath.Clean(r.URL.Path))

	// Return a 404 if the template doesn't exist
	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
	}

	// Return a 404 if the request is for a directory
	if info.IsDir() {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		// Log the detailed error
		log.Println(err.Error())
		// Return a generic "Internal Server Error" message
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", nil); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

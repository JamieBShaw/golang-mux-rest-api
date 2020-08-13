package handlers

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/JamieBShaw/golang-basic-microservice/products-images/files"

	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
)

// Files is a handler for reading and writing files
type Files struct {
	log   hclog.Logger
	store files.Storage
}

// NewFiles creates a new File handler
func NewFiles(s files.Storage, l hclog.Logger) *Files {
	return &Files{store: s, log: l}
}

// UploadREST implements the http.Handler interface
func (f *Files) UploadREST(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	fn := vars["filename"]

	f.log.Info("Handle POST", "id", id, "filename", fn)

	if id == "" || fn == "" {
		f.invalidURI(r.URL.String(), rw)
	}

	// no need to check for invalid id or filename as the mux router will not send requests
	// here unless they have the correct parameters

	f.saveFile(id, fn, rw, r.Body)
}

// UploadMultipart: Grabbing data from a multi input form
func (f *Files) UploadMultipart(rw http.ResponseWriter, r *http.Request) {

	err := r.ParseMultipartForm(128 * 1024)

	if err != nil {
		f.log.Error("BAD REQUEST", err)
		http.Error(rw, "Expected multipart form data", http.StatusBadRequest)
		return
	}

	id, idErr := strconv.Atoi(r.FormValue("id")) // Grabbing the form html element with name id
	f.log.Info("Process form for id", "id", id)

	if idErr != nil {
		f.log.Error("BAD REQUEST", err)
		http.Error(rw, "Expected integer id", http.StatusBadRequest)
		return
	}

	ff, mh, err := r.FormFile("file") // Grabbing the form html element with name file (image upload)
	if err != nil {
		f.log.Error("BAD REQUEST", err)
		http.Error(rw, "Expected file", http.StatusBadRequest)
		return
	}

	f.saveFile(r.FormValue("id"), mh.Filename, rw, ff)

}

func (f *Files) invalidURI(uri string, rw http.ResponseWriter) {
	f.log.Error("Invalid path", "path", uri)
	http.Error(rw, "Invalid file path should be in the format: /[id]/[filepath]", http.StatusBadRequest)
}

// saveFile saves the contents of the request to a file
func (f *Files) saveFile(id, path string, rw http.ResponseWriter, r io.ReadCloser) {
	f.log.Info("Save file for product", "id", id, "path", path)

	fp := filepath.Join(id, path)
	err := f.store.Save(fp, r)
	if err != nil {
		f.log.Error("Unable to save file", "error", err)
		http.Error(rw, "Unable to save file", http.StatusInternalServerError)
	}
}

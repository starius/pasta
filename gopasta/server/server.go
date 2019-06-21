package server

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/robfig/humanize"
	"github.com/starius/pasta/gopasta/database"
)

//go:generate go get github.com/jteeuwen/go-bindata/go-bindata
//go:generate go-bindata -pkg server -nometadata -nomemcopy -nocompress -prefix $GOPATH/src/github.com/starius/pasta/gopasta/server -o $GOPATH/src/github.com/starius/pasta/gopasta/server/favicon.go $GOPATH/src/github.com/starius/pasta/gopasta/server/favicon.ico
//go:generate go fmt $GOPATH/src/github.com/starius/pasta/gopasta/server/favicon.go

type IDEncoder interface {
	Encode(id uint64, longID bool) (phrase string, err error)
	Decode(phrase string) (id uint64, longID bool, err error)
}

type Handler struct {
	db        *database.Database
	idEncoder IDEncoder

	mux *http.ServeMux

	maxSize int64

	adminAuth string
}

func NewHandler(db *database.Database, idEncoder IDEncoder, maxSize int, adminAuth string) *Handler {
	faviconReader := bytes.NewReader(MustAsset("favicon.ico"))
	faviconHandler := func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "favicon.ico", time.Unix(0, 0), faviconReader)
	}

	h := &Handler{db, idEncoder, http.NewServeMux(), int64(maxSize), adminAuth}
	h.mux.HandleFunc("/favicon.ico", faviconHandler)
	h.mux.HandleFunc("/api/create", h.handleUpload)
	h.mux.HandleFunc("/", h.handleRecord)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("handleUpload was called with method %q.", r.Method)
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)
	if err := r.ParseMultipartForm(h.maxSize); err != nil {
		log.Printf("ParseMultipartForm: %v", err)
		http.Error(w, "failed to parse the form", http.StatusBadRequest)
		return
	}
	content := []byte(r.FormValue("content"))
	filename := r.FormValue("filename")
	selfBurning := r.FormValue("self_burning") != ""
	redirect := r.FormValue("redirect") != ""
	longID := r.FormValue("long_id") != ""
	if file, header, err := r.FormFile("file"); err == nil {
		if len(content) != 0 || filename != "" {
			log.Printf("Provided both text content and file")
			http.Error(w, "Provided both text content and file", http.StatusBadRequest)
			return
		}
		content, err = ioutil.ReadAll(file)
		if err != nil {
			log.Printf("Reading multipart file: %v.", err)
			http.Error(w, "Failed to read the file", http.StatusBadRequest)
			return
		}
		filename = header.Filename
	}
	record := &database.Record{
		Content:     content,
		Filename:    filename,
		SelfBurning: selfBurning,
		Redirect:    redirect,
		LongId:      longID,
	}
	id, err := h.db.Add(record)
	if err != nil {
		log.Printf("Add: %v.", err)
		http.Error(w, "failed to save the record", http.StatusInternalServerError)
		return
	}
	phrase, err := h.idEncoder.Encode(id, longID)
	if err != nil {
		log.Printf("idEncoder.Encode: %v.", err)
		http.Error(w, "failed to create the link", http.StatusInternalServerError)
		return
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	targetURL := fmt.Sprintf("%s://%s/%s", scheme, r.Host, phrase)
	if !selfBurning && !redirect {
		http.Redirect(w, r, targetURL, http.StatusFound)
	}
	fmt.Fprintf(w, "Your link: %s\n", targetURL)
}

func (h *Handler) handleRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodDelete {
		log.Printf("handleRecord was called with method %q.", r.Method)
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}
	if r.Method == http.MethodGet && (r.URL.Path == "" || r.URL.Path == "/") {
		h.handleMain(w, r)
		return
	}
	if r.Method == http.MethodDelete {
		auth := r.Header.Get("Authorization")
		if subtle.ConstantTimeCompare([]byte(auth), []byte(h.adminAuth)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}
	pathParts := strings.Split(r.URL.Path, "/")
	phrase := pathParts[len(pathParts)-1]
	if phrase == "" {
		if len(pathParts) == 1 {
			log.Printf("Bad URL: %s.", r.URL.Path)
			http.Error(w, "bad link", http.StatusBadRequest)
			return
		}
		phrase = pathParts[len(pathParts)-2]
	}
	id, longID, err := h.idEncoder.Decode(phrase)
	if err != nil {
		log.Printf("idEncoder.Decode(%q): %v.", phrase, err)
		http.Error(w, "bad link", http.StatusBadRequest)
		return
	}
	if r.Method == http.MethodDelete {
		if err := h.db.Delete(id); err != nil {
			log.Printf("db.Delete(%d): %v.", id, err)
			http.Error(w, "failed to delete the link", http.StatusInternalServerError)
		}
		return
	}
	record, err := h.db.Lookup(id)
	if err != nil {
		log.Printf("db.Lookup(%d): %v.", id, err)
		http.Error(w, "bad link", http.StatusBadRequest)
		return
	}
	if record.LongId != longID {
		log.Printf("long ID mismatch in %d.", id)
		http.Error(w, "bad link", http.StatusBadRequest)
		return
	}
	if record.Redirect {
		status := http.StatusMovedPermanently
		if record.SelfBurning {
			status = http.StatusFound
		}
		http.Redirect(w, r, string(bytes.TrimSpace(record.Content)), status)
	} else {
		w.Write(record.Content)
	}
}

func (h *Handler) handleMain(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		MaxSize string
		Uploads string
	}{
		MaxSize: humanize.IBytes(uint64(h.maxSize)),
		Uploads: humanize.Comma(h.db.RecordsCount()),
	}
	mainTemplate.Execute(w, vars)
}

package server

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/robfig/humanize"
	"github.com/starius/pasta/gopasta/database"
	"golang.org/x/net/idna"
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
		w.Header().Set("Content-Type", "image/vnd.microsoft.icon")
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

var punycode = idna.New()

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
	if redirect {
		content = bytes.TrimSpace(content)
		u, err := url.Parse(string(content))
		if err != nil {
			log.Printf("URL is invalid: %v.", err)
			http.Error(w, "URL is invalid", http.StatusBadRequest)
			return
		}
		if !u.IsAbs() {
			log.Printf("URL is not absolute.")
			http.Error(w, "URL is  not absolute", http.StatusBadRequest)
			return
		}
		content = []byte(u.String())
	}
	ctype := ""
	if !redirect {
		ctype = mime.TypeByExtension(filepath.Ext(filename))
		if ctype == "" {
			ctype = http.DetectContentType(content)
		}
	}
	record := &database.Record{
		Content:     content,
		Filename:    filename,
		SelfBurning: selfBurning,
		Redirect:    redirect,
		LongId:      longID,
		ContentType: ctype,
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
	host, err := punycode.ToUnicode(r.Host)
	if err != nil {
		log.Printf("punycode.ToUnicode(%q): %v.", r.Host, err)
		host = r.Host
	}
	targetURL := fmt.Sprintf("%s://%s/%s", scheme, host, phrase)
	if !selfBurning && !redirect {
		http.Redirect(w, r, targetURL, http.StatusFound)
	}
	fmt.Fprintf(w, "Your link: %s\n", targetURL)
}

func (h *Handler) handleRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodDelete {
		log.Printf("handleRecord was called with method %q.", r.Method)
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}
	if (r.Method == http.MethodGet || r.Method == http.MethodHead) && (r.URL.Path == "" || r.URL.Path == "/") {
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
		http.Redirect(w, r, string(record.Content), status)
	} else {
		if record.Filename != "" {
			w.Header().Add("Content-Disposition", fmt.Sprintf("inline; filename=%q", record.Filename))
		}
		if record.ContentType != "" {
			w.Header().Add("Content-Type", record.ContentType)
		}
		contentReader := bytes.NewReader(record.Content)
		http.ServeContent(w, r, record.Filename, time.Unix(0, 0), contentReader)
	}
}

func (h *Handler) handleMain(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		FileTab bool
		MaxSize string
		Uploads string
	}{
		FileTab: r.FormValue("filetab") == "on",
		MaxSize: humanize.IBytes(uint64(h.maxSize)),
		Uploads: humanize.Comma(h.db.RecordsCount()),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := mainTemplate.Execute(w, vars); err != nil {
		log.Printf("failed to execute template: %v", err)
	}
}

package server

import (
	"bytes"
	"crypto/subtle"
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/robfig/humanize"
	"github.com/starius/pasta/gopasta/database"
	"github.com/starius/pasta/src/static/pasta"
	"golang.org/x/net/idna"
)

//go:embed *.html
var templatesFs embed.FS

type IDEncoder interface {
	Encode(id uint64, longID bool) (phrase string, err error)
	Decode(phrase string) (id uint64, longID bool, err error)
}

type Handler struct {
	db         *database.Database
	idEncoder  IDEncoder
	mux        *http.ServeMux
	mainTmpl   *template.Template
	uploadTmpl *template.Template
	maxSize    int64
	adminAuth  string
	domains    []string
	allowFiles bool
	filesBurn  bool
	allBurn    bool

	filesScheme, filesHost string
}

func NewHandler(db *database.Database, idEncoder IDEncoder, maxSize int, adminAuth string, domains []string, allowFiles, filesBurn, allBurn bool, filesScheme, filesHost string) *Handler {
	faviconReader := bytes.NewReader(pasta.FaviconBytes)
	faviconHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/vnd.microsoft.icon")
		http.ServeContent(w, r, "favicon.ico", time.Unix(0, 0), faviconReader)
	}

	h := &Handler{
		db:          db,
		idEncoder:   idEncoder,
		mux:         http.NewServeMux(),
		mainTmpl:    template.Must(template.ParseFS(templatesFs, "index.html", "base.html")),
		uploadTmpl:  template.Must(template.ParseFS(templatesFs, "upload.html", "base.html")),
		maxSize:     int64(maxSize),
		adminAuth:   adminAuth,
		domains:     domains,
		allowFiles:  allowFiles,
		filesBurn:   filesBurn,
		allBurn:     allBurn,
		filesScheme: filesScheme,
		filesHost:   filesHost,
	}

	h.mux.HandleFunc("/favicon.ico", faviconHandler)
	h.mux.HandleFunc("/api/create", h.handleUpload)
	h.mux.HandleFunc("/", h.handleRecord)

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	h.mux.ServeHTTP(w, r)
}

var punycode = idna.New()

const textContentType = "text/plain; charset=utf-8"

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

	defer r.MultipartForm.RemoveAll()

	content := []byte(r.PostFormValue("content"))
	filename := r.PostFormValue("filename")
	selfBurning := r.PostFormValue("self_burning") != ""
	redirect := r.PostFormValue("redirect") != ""
	longID := r.PostFormValue("long_id") != ""

	if file, header, err := r.FormFile("file"); err == nil {
		defer file.Close()
		if len(content) != 0 || filename != "" {
			log.Printf("Provided both text content and file")
			http.Error(w, "Provided both text content and file", http.StatusBadRequest)
			return
		}
		content, err = io.ReadAll(file)
		if err != nil {
			log.Printf("Reading multipart file: %v.", err)
			http.Error(w, "Failed to read the file", http.StatusBadRequest)
			return
		}
		filename = header.Filename
	}

	if len(filename) > 255 || strings.IndexFunc(filename, func(r rune) bool { return r < 0x20 || r == 0x7f }) >= 0 {
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}
	if int64(len(content)) > h.maxSize {
		http.Error(w, "content is too large", http.StatusRequestEntityTooLarge)
		return
	}
	if redirect {
		content = bytes.TrimSpace(content)
		if !validRedirectURL(string(content)) {
			http.Error(w, "URL must use http or https with a host", http.StatusBadRequest)
			return
		}
		if len(content) > 1024 {
			http.Error(w, "URL is too long", http.StatusBadRequest)
			return
		}
	}
	if len(content) == 0 {
		log.Printf("Empty content.")
		http.Error(w, "Empty content.", http.StatusBadRequest)
		return
	}

	ctype := ""
	if !redirect {
		ctype = mime.TypeByExtension(filepath.Ext(filename))
		if ctype == "" {
			ctype = http.DetectContentType(content)
		}
	}

	var validText bool
	if !h.allowFiles || h.filesBurn {
		// Check of the content is needed.
		validText = utf8.ValidString(string(content))
	}
	if !h.allowFiles && !validText {
		log.Printf("Invalid content.")
		http.Error(w, "Invalid content.", http.StatusBadRequest)
		return
	}
	if h.filesBurn && !validText {
		// In filesBurn mode all files are forced self-burning.
		selfBurning = true
	}
	if h.allBurn {
		selfBurning = true
	}
	if h.filesBurn && !selfBurning {
		ctype = textContentType
	}
	isFile := !redirect && ctype != textContentType

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
	if h.filesHost != "" && isFile {
		scheme = h.filesScheme
		host = h.filesHost
	}
	targetURL := fmt.Sprintf("%s://%s/%s", scheme, host, phrase)
	if !selfBurning && !redirect && contentTypeIsInline(ctype) {
		http.Redirect(w, r, targetURL, http.StatusFound)
	}

	if r.PostFormValue("browser") != "on" {
		fmt.Fprintf(w, "Your link: %s\n", targetURL)
		return
	}

	vars := struct {
		URL         string
		HumanURL    string
		ID          string
		SelfBurning bool
	}{
		URL:         targetURL,
		HumanURL:    fmt.Sprintf("%s/%s", host, phrase),
		ID:          phrase,
		SelfBurning: selfBurning,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.uploadTmpl.Execute(w, vars); err != nil {
		log.Printf("failed to execute upload template: %v", err)
	}
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
			http.Error(w, "bad link", http.StatusBadRequest)
			return
		}
		phrase = pathParts[len(pathParts)-2]
	}
	if len(phrase) > 256 {
		http.Error(w, "bad link", http.StatusBadRequest)
		return
	}
	id, longID, err := h.idEncoder.Decode(phrase)
	if err != nil {
		log.Printf("invalid record ID")
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

	// In case of file download, DuckDuckGO sends two requests: first
	// request is a probe, sent by WebView component of DuckDuckGO. Its
	// result is not saved. Then another request is sent by the browser
	// itself and its result is saved. To prevent a self-burning file from
	// expiring after the first request, we detect the first request by
	// header X-Requested-With and keep the file, returning zero bytes in
	// response. The length of response is preserved though. For additional
	// context see https://github.com/duckduckgo/Android/issues/5164
	// Similar problem exists in Telegram embedded browser.
	xrw := r.Header.Get("X-Requested-With")
	probe := r.Method == http.MethodHead || strings.Contains(xrw, "duckduckgo") || strings.Contains(xrw, "telegram")

	record, err := h.db.LookupValidated(id, r.UserAgent(), probe, func(record *database.Record) error {
		if record.LongId != longID {
			return fmt.Errorf("record ID type mismatch")
		}
		if record.Redirect {
			if !validRedirectURL(string(record.Content)) {
				return fmt.Errorf("invalid redirect URL")
			}
		} else if record.ContentType != textContentType && h.filesHost != "" && r.Host != h.filesHost {
			return fmt.Errorf("incorrect file download host")
		}
		return nil
	})
	if err != nil {
		log.Printf("db.Lookup(%d): %v.", id, err)
		http.Error(w, "bad link", http.StatusBadRequest)
		return
	}
	if record.Redirect {
		if probe {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		status := http.StatusMovedPermanently
		if record.SelfBurning {
			status = http.StatusFound
		}
		http.Redirect(w, r, string(record.Content), status)
	} else {
		// User uploads must never gain the site's origin or execute active content.
		w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'; base-uri 'none'; form-action 'none'")
		disposition := "attachment"
		if contentTypeIsInline(record.ContentType) {
			disposition = "inline"
		}
		params := map[string]string{}
		if record.Filename != "" {
			params["filename"] = record.Filename
		}
		w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, params))
		ctype := record.ContentType
		if ctype == "" {
			ctype = "application/octet-stream"
		}
		w.Header().Set("Content-Type", ctype)
		contentReader := bytes.NewReader(record.Content)
		http.ServeContent(w, r, record.Filename, time.Unix(0, 0), contentReader)
	}
}

func (h *Handler) handleMain(w http.ResponseWriter, r *http.Request) {
	vars := struct {
		TextTab     bool
		AllowFiles  bool
		FileTab     bool
		ShortnerTab bool
		ForcedBurn  bool
		MaxSize     string
		Uploads     string
		Domains     []string
	}{
		TextTab:     r.FormValue("tab") == "text" || r.FormValue("tab") == "",
		AllowFiles:  h.allowFiles,
		FileTab:     r.FormValue("tab") == "file",
		ShortnerTab: r.FormValue("tab") == "shortner",
		MaxSize:     humanize.IBytes(uint64(h.maxSize)),
		Uploads:     humanize.Comma(h.db.RecordsCount()),
		Domains:     h.domains,
	}
	vars.ForcedBurn = h.filesBurn && vars.FileTab || h.allBurn
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.mainTmpl.Execute(w, vars); err != nil {
		log.Printf("failed to execute template: %v", err)
	}
}

// validRedirectURL also checks stored records, which may predate upload validation.
func validRedirectURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != "" && u.Opaque == ""
}

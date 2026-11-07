package server

import (
	"bytes"
	"crypto/aes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/starius/pasta/gopasta/database"
)

type testIDEncoder struct{}

func (testIDEncoder) Encode(id uint64, longID bool) (string, error) {
	prefix := "short-"
	if longID {
		prefix = "long-"
	}
	return prefix + strconv.FormatUint(id, 10), nil
}

func (testIDEncoder) Decode(value string) (uint64, bool, error) {
	prefix, number, ok := strings.Cut(value, "-")
	if !ok || (prefix != "short" && prefix != "long") {
		return 0, false, fmt.Errorf("bad ID")
	}
	id, err := strconv.ParseUint(number, 10, 64)
	return id, prefix == "long", err
}

func testHandler(t *testing.T) *Handler {
	t.Helper()
	dir := t.TempDir()
	index, err := os.Create(filepath.Join(dir, "index"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { index.Close() })
	data, err := os.Create(filepath.Join(dir, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { data.Close() })
	indexBlock, _ := aes.NewCipher(make([]byte, 16))
	dataBlock, _ := aes.NewCipher(bytes.Repeat([]byte{1}, 16))
	const maxSize = 1 << 20
	db, err := database.NewDatabase(index, data, indexBlock, dataBlock, maxSize, 10, maxSize, false)
	if err != nil {
		t.Fatal(err)
	}
	return NewHandler(db, testIDEncoder{}, maxSize, "admin-token", nil, true, false, false, "", "")
}

func addRecord(t *testing.T, h *Handler, record *database.Record) string {
	t.Helper()
	id, err := h.db.Add(record)
	if err != nil {
		t.Fatal(err)
	}
	phrase, err := h.idEncoder.Encode(id, record.LongId)
	if err != nil {
		t.Fatal(err)
	}
	return "http://example.test/" + phrase
}

func request(h *Handler, method, target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(method, target, nil))
	return w
}

func upload(t *testing.T, h *Handler, target string, values map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	for key, value := range values {
		if err := form.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, target, &body)
	r.Header.Set("Content-Type", form.FormDataContentType())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestUploadedContentIsolation(t *testing.T) {
	for _, ctype := range []string{"text/html", "image/svg+xml", "application/javascript", "text/javascript", "application/xhtml+xml", "text/css", "image/png-attacker"} {
		t.Run(ctype, func(t *testing.T) {
			h := testHandler(t)
			target := addRecord(t, h, &database.Record{Content: []byte("<script>alert(document.domain)</script>"), Filename: "evil.html", ContentType: ctype})
			w := request(h, http.MethodGet, target)
			if w.Code != http.StatusOK || !strings.HasPrefix(w.Header().Get("Content-Disposition"), "attachment") {
				t.Fatalf("unsafe response: %d %v", w.Code, w.Header())
			}
			if !strings.Contains(w.Header().Get("Content-Security-Policy"), "sandbox") {
				t.Fatal("missing sandbox")
			}
			if w.Header().Get("X-Content-Type-Options") != "nosniff" || w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("Referrer-Policy") != "no-referrer" {
				t.Fatalf("missing privacy headers: %v", w.Header())
			}
		})
	}
	h := testHandler(t)
	target := addRecord(t, h, &database.Record{Content: []byte("plain text"), ContentType: textContentType})
	if got := request(h, http.MethodGet, target).Header().Get("Content-Disposition"); got != "inline" {
		t.Fatalf("plain text disposition = %q", got)
	}
}

func TestUnsafeRedirectsRejected(t *testing.T) {
	for _, target := range []string{"javascript:alert(1)", "data:text/html,<script>alert(1)</script>", "file:///etc/passwd", "http:example.test", "https:///path", "//example.test", "https://"} {
		t.Run(target, func(t *testing.T) {
			h := testHandler(t)
			w := upload(t, h, "http://example.test/api/create", map[string]string{"content": target, "redirect": "on"})
			if w.Code != http.StatusBadRequest || h.db.RecordsCount() != 0 {
				t.Fatalf("unsafe upload accepted: %d", w.Code)
			}
			stored := addRecord(t, h, &database.Record{Content: []byte(target), Redirect: true})
			w = request(h, http.MethodGet, stored)
			if w.Code != http.StatusBadRequest || w.Header().Get("Location") != "" {
				t.Fatalf("unsafe stored redirect accepted: %d %v", w.Code, w.Header())
			}
		})
	}
	for _, target := range []string{"http://example.test/path?q=x", "https://example.test/path"} {
		h := testHandler(t)
		stored := addRecord(t, h, &database.Record{Content: []byte(target), Redirect: true})
		w := request(h, http.MethodGet, stored)
		if w.Code != http.StatusMovedPermanently || w.Header().Get("Location") != target {
			t.Fatalf("valid redirect rejected: %d %v", w.Code, w.Header())
		}
	}
}

func TestUploadRejectsUnsafeFilenamesAndQueryContent(t *testing.T) {
	h := testHandler(t)
	for _, filename := range []string{strings.Repeat("x", 256), "file\r\nInjected: header", "file\x00.txt"} {
		w := upload(t, h, "http://example.test/api/create", map[string]string{"content": "text", "filename": filename})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("filename %q accepted: %d", filename, w.Code)
		}
	}
	w := upload(t, h, "http://example.test/api/create?content=query-secret", nil)
	if w.Code != http.StatusBadRequest || h.db.RecordsCount() != 0 {
		t.Fatalf("query content accepted: %d", w.Code)
	}
}

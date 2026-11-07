package main

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mnemonics "gitlab.com/NebulousLabs/entropy-mnemonics"
)

func TestDecodeSecret(t *testing.T) {
	want := bytes.Repeat([]byte{42}, 16)
	phrase, err := mnemonics.ToPhrase(want, mnemonics.English)
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeSecret([]byte(phrase.String() + "\n"))
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("generated phrase did not round trip: %x, %v", got, err)
	}
	for _, secret := range []string{"", "  ", "abbey", "abbey abbey", "unknown", strings.Replace(phrase.String(), " ", "  ", 1), "\xff"} {
		if _, err := decodeSecret([]byte(secret)); err == nil {
			t.Errorf("accepted invalid or short secret %q", secret)
		}
	}
}

func TestConcurrentRequestLimit(t *testing.T) {
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	handler := limitRequests(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entered <- struct{}{}
		<-release
		w.WriteHeader(http.StatusNoContent)
	}), 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	}()
	<-entered
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
	if response.Code != http.StatusServiceUnavailable || response.Header().Get("Retry-After") == "" {
		t.Errorf("saturated handler returned %d, %v", response.Code, response.Header())
	}
	close(release)
	<-done
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
	if response.Code != http.StatusNoContent {
		t.Errorf("request slot was not released: %d", response.Code)
	}
}

func TestServerRejectsStalledRequestBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			http.Error(w, "request timed out", http.StatusRequestTimeout)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	srv := newHTTPServer("127.0.0.1:0", handler, 50*time.Millisecond, time.Second)
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	go srv.Serve(ln)
	conn, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	io.WriteString(conn, "POST / HTTP/1.1\r\nHost: localhost\r\nContent-Length: 100\r\nConnection: close\r\n\r\nx")
	response, err := io.ReadAll(conn)
	if err != nil || !bytes.Contains(response, []byte("408 Request Timeout")) {
		t.Fatalf("stalled body was not rejected: %s, %v", response, err)
	}
}

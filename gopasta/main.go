package main

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/starius/pasta/gopasta/database"
	"github.com/starius/pasta/gopasta/server"
	"github.com/tyler-smith/go-bip39/wordlists"
	mnemonics "gitlab.com/NebulousLabs/entropy-mnemonics"
	"gitlab.com/starius/encrypt-autocert-cache"
	"gitlab.com/starius/fpe/phrase"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/term"
)

var (
	dir          = flag.String("dir", ".", "Directory to store data in")
	listen       = flag.String("listen", "127.0.0.1:8042", "Address for the plaintext HTTP listener")
	letsDomains  = flag.String("letsencrypt-domains", "", "Run Let's Encrypt on these domains")
	domains      = flag.String("domains", "", "List of domains listed on the main page")
	secretFile   = flag.String("secret-file", "", "Secret file")
	genSecret    = flag.Bool("gen-secret", false, "Generate random and exit")
	printAdmin   = flag.Bool("print-admin-auth", false, "Print admin Authorization header and exit")
	allowFiles   = flag.Bool("allow-files", true, "Allow binary files, not only texts")
	filesBurn    = flag.Bool("files-burn", false, "Binary files are forced to be self-burning")
	allBurn      = flag.Bool("all-burn", false, "All records are forced to be self-burning")
	filesScheme  = flag.String("files-scheme", "", "If set, downloading of files is available only using this URL scheme")
	filesHost    = flag.String("files-host", "", "If set, downloading of files is available only using this hostname")
	maxSize      = flag.Int("max-size", 10*1024*1024, "Max record size, bytes")
	cacheRecords = flag.Int("cache-records", 10000, "Cache size, records")
	cacheBytes   = flag.Int("cache-bytes", 100*1024*1024, "Cache size, bytes")
	readTimeout  = flag.Duration("read-timeout", 2*time.Minute, "Maximum time to read a request, including its body")
	writeTimeout = flag.Duration("write-timeout", 2*time.Minute, "Maximum time to write a response")
	maxRequests  = flag.Int("max-requests", 16, "Maximum concurrent application requests across listeners")
)

const saltHex = "b59e698ae2b5893a2a45edf3f809ef5977aa9b3526fbb76cf188817d6bbf19e3"
const databaseInfo = "database"
const idInfo = "id"
const adminAuthInfo = "adminAuth"
const certInfo = "cert"

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	url := "https://" + r.Host + r.URL.EscapedPath()
	if len(r.URL.RawQuery) > 0 {
		url += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, url, http.StatusPermanentRedirect)
}

func decodeSecret(secret []byte) ([]byte, error) {
	secret = bytes.TrimSpace(secret)
	// The mnemonic decoder treats empty words as valid words. Do not silently
	// normalize them, since that would derive a different key for existing data.
	if !utf8.Valid(secret) || len(secret) == 0 || string(secret) != strings.Join(strings.Fields(string(secret)), " ") {
		return nil, fmt.Errorf("invalid secret phrase; use a phrase produced by -gen-secret")
	}
	key, err := mnemonics.FromString(string(secret), mnemonics.English)
	if err != nil || len(key) < 16 {
		return nil, fmt.Errorf("secret must encode at least 16 bytes; use -gen-secret")
	}
	return key, nil
}

func newHTTPServer(addr string, handler http.Handler, readTimeout, writeTimeout time.Duration) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       time.Minute,
		MaxHeaderBytes:    32 << 10,
	}
}

func limitRequests(handler http.Handler, maximum int) http.Handler {
	slots := make(chan struct{}, maximum)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
			handler.ServeHTTP(w, r)
		default:
			w.Header().Set("Retry-After", "1")
			w.Header().Set("Cache-Control", "no-store")
			http.Error(w, "server busy", http.StatusServiceUnavailable)
		}
	})
}

func main() {
	flag.Parse()
	if *genSecret {
		secret := make([]byte, 16)
		if _, err := rand.Read(secret); err != nil {
			log.Fatal("Failed to generate secret")
		}
		p, err := mnemonics.ToPhrase(secret, mnemonics.English)
		if err != nil {
			panic(err)
		}
		fmt.Println(p.String())
		os.Exit(0)
	}
	if *dir == "" {
		log.Fatalf("Please specify -dir")
	}
	if *maxSize <= 0 || *cacheRecords < 0 || *cacheBytes < 0 || *maxRequests <= 0 || *readTimeout <= 0 || *writeTimeout <= 0 {
		log.Fatal("Size, request and timeout limits must be positive; cache limits may be zero")
	}
	var err error
	var secret []byte
	if *secretFile != "" {
		secret, err = os.ReadFile(*secretFile)
		if err != nil {
			log.Fatalf("Failed to read secret file")
		}
	} else {
		fmt.Print("Enter secret: ")
		secret, err = term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			log.Fatalf("Failed to read secret from keyboard")
		}
	}
	mainKey, err := decodeSecret(secret)
	if err != nil {
		log.Fatalf("Failed to decode the secret: %v", err)
	}
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		panic(err)
	}

	adminAuthHKDF := hkdf.New(sha256.New, mainKey, salt, []byte(adminAuthInfo))
	adminAuthBytes := make([]byte, 16)
	if _, err := io.ReadFull(adminAuthHKDF, adminAuthBytes); err != nil {
		panic(err)
	}
	adminAuth := hex.EncodeToString(adminAuthBytes)
	if *printAdmin {
		fmt.Printf("curl -X DELETE -H 'Authorization: %s' URL-to-delete\n", adminAuth)
		os.Exit(0)
	}

	dbHKDF := hkdf.New(sha256.New, mainKey, salt, []byte(databaseInfo))
	indexKey := make([]byte, 16)
	if _, err := io.ReadFull(dbHKDF, indexKey); err != nil {
		panic(err)
	}
	dataKey := make([]byte, 16)
	if _, err := io.ReadFull(dbHKDF, dataKey); err != nil {
		panic(err)
	}
	index, err := os.OpenFile(*dir+"/index", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	data, err := os.OpenFile(*dir+"/data", os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	indexBlock, err := aes.NewCipher(indexKey)
	if err != nil {
		panic(err)
	}
	dataBlock, err := aes.NewCipher(dataKey)
	if err != nil {
		panic(err)
	}
	db, err := database.NewDatabase(index, data, indexBlock, dataBlock, *maxSize, *cacheRecords, *cacheBytes, *allBurn)
	if err != nil {
		panic(err)
	}

	idHKDF := hkdf.New(sha256.New, mainKey, salt, []byte(idInfo))
	wordsKey := make([]byte, 16)
	if _, err := io.ReadFull(idHKDF, wordsKey); err != nil {
		panic(err)
	}
	longKey := make([]byte, 16)
	if _, err := io.ReadFull(idHKDF, longKey); err != nil {
		panic(err)
	}
	wordsBlock, err := aes.NewCipher(wordsKey)
	if err != nil {
		panic(err)
	}
	longBlock, err := aes.NewCipher(longKey)
	if err != nil {
		panic(err)
	}
	idEncoder, err := phrase.NewIDEncoder(wordsBlock, longBlock, wordlists.English)
	if err != nil {
		panic(err)
	}

	domainsList := strings.Split(*domains, ",")
	if len(domainsList[0]) == 0 {
		domainsList = nil
	}
	handler := limitRequests(server.NewHandler(db, idEncoder, *maxSize, adminAuth, domainsList, *allowFiles, *filesBurn, *allBurn, *filesScheme, *filesHost), *maxRequests)

	if *letsDomains != "" {
		// Use Let's Encrypt.
		letsDomainsList := strings.Split(*letsDomains, ",")

		// Encrypted cert cache.
		certHKDF := hkdf.New(sha256.New, mainKey, salt, []byte(certInfo))
		certKey := make([]byte, 32)
		if _, err := io.ReadFull(certHKDF, certKey); err != nil {
			panic(err)
		}
		cache, err := encrypt.NewEncryptedCache(autocert.DirCache(*dir), certKey)
		if err != nil {
			panic(err)
		}

		m := &autocert.Manager{
			Cache:      cache,
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(letsDomainsList...),
		}
		certHandler := m.HTTPHandler(http.HandlerFunc(redirectToHTTPS))
		tlsConfig := m.TLSConfig()
		tlsConfig.MinVersion = tls.VersionTLS12
		tcpLn, err := net.Listen("tcp", ":443")
		if err != nil {
			log.Fatal(err)
		}
		ln, err := encrypt.NewListener(tlsConfig, tcpLn)
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			log.Fatal(newHTTPServer(":80", certHandler, *readTimeout, *writeTimeout).ListenAndServe())
		}()
		go func() {
			log.Fatal(newHTTPServer(":443", handler, *readTimeout, *writeTimeout).Serve(ln))
		}()
	}

	log.Fatal(newHTTPServer(*listen, handler, *readTimeout, *writeTimeout).ListenAndServe())
}

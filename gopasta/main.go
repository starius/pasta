package main

import (
	"bytes"
	"crypto/aes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/starius/pasta/gopasta/database"
	"github.com/starius/pasta/gopasta/server"
	"github.com/tyler-smith/go-bip39/wordlists"
	"gitlab.com/NebulousLabs/entropy-mnemonics"
	"gitlab.com/NebulousLabs/fastrand"
	"gitlab.com/starius/encrypt-autocert-cache"
	"gitlab.com/starius/fpe/phrase"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	dir          = flag.String("dir", ".", "Directory to store data in")
	listen       = flag.String("listen", ":8042", "Address to listen on")
	letsDomains  = flag.String("letsencrypt-domains", "", "Run Let's Encrypt on these domains")
	domains      = flag.String("domains", "", "List of domains listed on the main page")
	secretFile   = flag.String("secret-file", "", "Secret file")
	genSecret    = flag.Bool("gen-secret", false, "Generate random and exit")
	printAdmin   = flag.Bool("print-admin-auth", false, "Print admin Authorization header and exit")
	allowFiles   = flag.Bool("allow-files", false, "Allow binary files, not only texts")
	maxSize      = flag.Int("max-size", 10*1024*1024, "Max record size, bytes")
	cacheRecords = flag.Int("cache-records", 10000, "Cache size, records")
	cacheBytes   = flag.Int("cache-bytes", 100*1024*1024, "Cache size, bytes")
)

const saltHex = "b59e698ae2b5893a2a45edf3f809ef5977aa9b3526fbb76cf188817d6bbf19e3"
const databaseInfo = "database"
const idInfo = "id"
const adminAuthInfo = "adminAuth"
const certInfo = "cert"

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	url := "https://" + r.Host + r.URL.Path
	if len(r.URL.RawQuery) > 0 {
		url += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, url, http.StatusPermanentRedirect)
}

func main() {
	flag.Parse()
	if *genSecret {
		secret := fastrand.Bytes(16)
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
	var err error
	var secret []byte
	if *secretFile != "" {
		secret, err = ioutil.ReadFile(*secretFile)
		if err != nil {
			log.Fatalf("Failed to read secret file")
		}
	} else {
		fmt.Print("Enter secret: ")
		secret, err = terminal.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			log.Fatalf("Failed to read secret from keyboard")
		}
	}
	secret = bytes.TrimSpace(secret)
	mainKey, err := mnemonics.FromString(string(secret), mnemonics.English)
	if err != nil {
		log.Fatalf("Failed to decode the secret")
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
		fmt.Println(adminAuth)
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
	db, err := database.NewDatabase(index, data, indexBlock, dataBlock, *maxSize, *cacheRecords, *cacheBytes)
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
	handler := server.NewHandler(db, idEncoder, *maxSize, adminAuth, domainsList, *allowFiles)

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
		listener := m.Listener()
		go func() {
			log.Fatal(http.ListenAndServe(":80", certHandler))
		}()
		go func() {
			log.Fatal(http.Serve(listener, handler))
		}()
	}

	log.Fatal(http.ListenAndServe(*listen, handler))
}

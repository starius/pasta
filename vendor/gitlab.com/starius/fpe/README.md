[![Godoc](https://godoc.org/gitlab.com/starius/fpe?status.svg)](https://godoc.org/gitlab.com/starius/fpe)

# fpe - Format-preserving encryption for numbers in Go

This package implements a block cipher operating on a domain of numbers
{0, 1, ..., maxvalue}, where maxvalue is an arbitrary number more than 1.

Usage:

```go
import "gitlab.com/starius/fpe"


blockCipher, _ := aes.NewCipher([]byte("0123456789abcdef"))
maxvalue := 100 // Numbers {0, 1, ..., 100}. Size of the domain: 101.
plaintext := 30
ciphertext := SimpleEncrypt(blockCipher, plaintext, maxvalue)
fmt.Println(ciphertext) // 21
plaintext = SimpleDecrypt(blockCipher, ciphertext, maxvalue)
fmt.Println(plaintext)  // 30
```

The implementation splits the bits of the number into two parts and runs 10
rounds of Feistel network on it, which is secure against all adaptative chosen
plaintext and chosen ciphertext attacks (CPCA) when m << 2^(n(1-ε)), where
m is the number of queries, according to [Jacques Patarin CRYPTO'03][crypto03].

The same article says that

 - for 4 rounds or more, a random Feistel scheme is secure against known plaintext attacks (KPA).
 - for 7 rounds or more it is secure against all adaptative chosen plaintext attacks (CPA).

If you want to reduce the number of rounds to 7 or 4, you can use functions
[Encrypt][encrypt] and [Decrypt][decrypt] which have the number of rounds as a parameter.
You can also specify "tweak" to those functions.

Performance (machine Intel(R) Xeon(R) CPU E5-1410 v2 @ 2.80GHz):

```
$ go test -bench . -run=BenchmarkEncrypt
goos: linux
goarch: amd64
pkg: gitlab.com/starius/fpe
BenchmarkEncrypt_30_100-8        3000000               511 ns/op           1.95 MB/s
BenchmarkEncrypt_8bytes-8        3000000               504 ns/op          15.86 MB/s
PASS
ok      gitlab.com/starius/fpe  4.079s
```

AES performance on the same machine as reference:

```
$ go test crypto/cipher -bench GCM
goos: linux
goarch: amd64
pkg: crypto/cipher
BenchmarkAESGCMSeal1K-8          1000000              1023 ns/op        1000.76 MB/s
BenchmarkAESGCMOpen1K-8          1000000              1011 ns/op        1012.17 MB/s
BenchmarkAESGCMSign8K-8           300000              5469 ns/op        1497.80 MB/s
BenchmarkAESGCMSeal8K-8           200000              7467 ns/op        1097.06 MB/s
BenchmarkAESGCMOpen8K-8           200000              7415 ns/op        1104.68 MB/s
PASS
ok      crypto/cipher   7.851s
```

[crypto03]: http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.702.6231
[encrypt]: https://godoc.org/gitlab.com/starius/fpe#Encrypt
[decrypt]: https://godoc.org/gitlab.com/starius/fpe#Decrypt

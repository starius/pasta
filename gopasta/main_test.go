package main

import (
	"bytes"
	mnemonics "gitlab.com/NebulousLabs/entropy-mnemonics"
	"strings"
	"testing"
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

package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"fmt"
	"github.com/deckarep/gosx-notifier"
	"github.com/havoc-io/go-keytar"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/pbkdf2"
)

const (
	salt       = "saltysalt"
	iterations = 1003
	keylength  = 16
)

func main() {
	db, err := sqlx.Open("sqlite3", "/Users/sereger/Library/Application Support/Google/Chrome/Default/Cookies")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	derivedKey, err := getDerivedKey()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	q := `select encrypted_value from cookies where host_key ="bitbucket.org" AND name = "JSESSIONID"`
	encryptedValue := []byte{}
	err = db.GetContext(ctx, &encryptedValue, q)
	if err != nil {
		panic(err)
	}

	val, err := chromeDecrypt(derivedKey, encryptedValue[3:])
	if err != nil {
		panic(err)
	}

	note := gosxnotifier.NewNotification(fmt.Sprintf("JSESSIONID = %s", val))
	note.Title = "jira info"
	note.Subtitle = "statistic ok"
	note.Push()
}

func chromeDecrypt(key []byte, encrypted []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	iv := make([]byte, 16)
	for i := 0; i < 16; i++ {
		iv[i] = ' '
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(encrypted))
	blockMode.CryptBlocks(origData, encrypted)
	origData = pkcs5UnPadding(origData)
	return string(origData), nil
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func getDerivedKey() ([]byte, error) {
	keychain, err := keytar.GetKeychain()
	if err != nil {
		return nil, fmt.Errorf("get chain: %w", err)
	}
	chromePassword, err := keychain.GetPassword("Chrome Safe Storage", "Chrome")
	if err != nil {
		return nil, fmt.Errorf("get chome pass: %w", err)
	}

	dk := pbkdf2.Key([]byte(chromePassword), []byte(salt), iterations, keylength, sha1.New)
	return dk, nil
}

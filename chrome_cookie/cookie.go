package chrome_cookie

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"

	"github.com/havoc-io/go-keytar"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/pbkdf2"
)

const (
	dbPath     = "/Library/Application Support/Google/Chrome/Default/Cookies"
	salt       = "saltysalt"
	iterations = 1003
	keylength  = 16
)

type CookieReader struct {
	dbPath    string
	chromeKey []byte
}

func NewCookieReader() (*CookieReader, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home: %w", err)
	}
	chromeKey, err := getDerivedKey()
	if err != nil {
		return nil, fmt.Errorf("get chrome key: %w", err)
	}

	return &CookieReader{
		dbPath:    home + dbPath,
		chromeKey: chromeKey,
	}, nil
}

func (cr *CookieReader) Cookies(ctx context.Context, domain string) (map[string]string, error) {
	if len(domain) == 0 {
		return nil, errors.New("domain not set")
	}

	db, err := sqlx.Open("sqlite3", cr.dbPath)
	if err != nil {
		return nil, fmt.Errorf("open chrome db: %w", err)
	}
	defer db.Close()

	q := `select name, encrypted_value from cookies where host_key Like "%` + domain + `"`
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("exec query: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	name, encValue := "", []byte{}
	for rows.Next() {
		err = rows.Scan(&name, &encValue)
		if err != nil {
			return nil, fmt.Errorf("read row from db: %w", err)
		}

		val, err := chromeDecrypt(cr.chromeKey, encValue[3:])
		if err != nil {
			fmt.Printf("cannot decript value for key [%s]: %s", name, err)
			continue
		}

		result[name] = val
	}

	return result, nil
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

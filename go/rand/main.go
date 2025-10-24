package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func randomByte(l int) []byte {
	output := make([]byte, l)
	rand.Read(output)
	return output
}

func encryption(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(plaintext))
	cipher.NewCFBEncrypter(block, iv).XORKeyStream(ciphertext, plaintext)
	return ciphertext, nil
}

var (
	key []byte
	iv  []byte
)

func selfCheck(ts int64, key []byte, iv []byte) error {
	rand.Seed(ts)
	newKey := randomByte(len(key))
	newIv := randomByte(len(iv))
	if !bytes.Equal(key, newKey) {
		return errors.New("key is not equal")
	}
	if !bytes.Equal(iv, newIv) {
		return errors.New("iv is not equal")
	}
	return nil
}

func main() {
	now := time.Now().Unix()
	rand.Seed(now)
	key = randomByte(16)
	iv = randomByte(aes.BlockSize)
	if err := selfCheck(now, key, iv); err != nil {
		fmt.Println("selfCheck failed:", err)
		return
	}

	e := echo.New()
	e.Debug = true
	e.Use(middleware.Gzip())
	e.GET("/", func(c echo.Context) error {
		flag := os.Getenv("flag")

		ciphertext, err := encryption([]byte(flag))
		if err != nil {
			return err
		}

		encodedCiphertext := base64.StdEncoding.EncodeToString(ciphertext)

		return c.HTML(http.StatusOK, fmt.Sprintf(`Ciphertext: %s`, encodedCiphertext))
	})

	e.Logger.Fatal(e.Start(":80"))
}

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
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

func main() {
	now := time.Now().Unix()
	rand.Seed(now)
	key = randomByte(16)
	iv = randomByte(aes.BlockSize)

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

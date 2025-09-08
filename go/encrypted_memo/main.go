package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

// 密钥变量 - 会在启动时从文件加载或随机生成
var key []byte

// generateRandomKey 生成一个32字节的随机密钥用于AES-256
func generateRandomKey() []byte {
	key := make([]byte, 32) // AES-256 需要32字节密钥
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		log.Fatalf("Failed to generate random key: %v", err)
	}
	return key
}

// init 函数在main函数之前执行，用于初始化密钥
func init() {
	key = generateRandomKey()
	log.Printf("Using encryption key (length: %d bytes)", len(key))
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("empty data")
	}
	unpadding := int(data[length-1])
	if unpadding > length || unpadding > blockSize {
		return nil, errors.New("invalid padding")
	}
	return data[:(length - unpadding)], nil
}

// --- Encryption/Decryption ---
func encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	plaintext = pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func decrypt(b64ciphertext string) (string, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(b64ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", errors.New("invalid padding")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	unpadded, err := pkcs7Unpad(ciphertext, aes.BlockSize)
	if err != nil {
		return "", errors.New("invalid padding")
	}
	return string(unpadded), nil
}

func main() {
	flag := os.Getenv("flag")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie("session")
		var sessionData string

		// If no cookie, create one for a guest user.
		if err != nil {
			userData := "user=guest&is_admin=false"
			encryptedSession, _ := encrypt([]byte(userData))
			http.SetCookie(w, &http.Cookie{Name: "session", Value: encryptedSession, Path: "/"})
			sessionData = userData
		} else {
			// If cookie exists, try to decrypt it.
			decrypted, err := decrypt(sessionCookie.Value)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			sessionData = decrypted
		}

		// Parse the session and check for admin rights.
		values, err := url.ParseQuery(sessionData)
		if err != nil {
			http.Error(w, "Error: Cannot parse session data", http.StatusBadRequest)
			return
		}

		user := values.Get("user")
		isAdmin := values.Get("is_admin")

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<h1>Go-pher的加密备忘录</h1>")
		if isAdmin == "true" {
			fmt.Fprintf(w, "<p>欢迎回来，<strong>管理员</strong>！</p>")
			fmt.Fprintf(w, "<p>这是您的秘密备忘录：<code>%s</code></p>", flag)
		} else {
			fmt.Fprintf(w, "<p>欢迎，%s！</p>", user)
			fmt.Fprintf(w, "<p>您没有任何备忘录。</p>")
		}
	})

	log.Println("Server starting on :80")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

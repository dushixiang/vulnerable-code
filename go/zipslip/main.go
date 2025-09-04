package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func writeFile(filename string, content []byte, perm os.FileMode) error {
	_, err := os.Stat(filename)
	if err == nil {
		_ = os.Remove(filename)
	}
	if err := os.WriteFile(filename, content, perm); err != nil {
		return err
	}
	return nil
}

func unzipUnsafe(dest string, data []byte) error {
	// Intentionally vulnerable: does not validate the canonical path resides under dest
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	for _, f := range zr.File {
		fullPath := filepath.Join(dest, f.Name)
		// Create parent directories as declared in the archive without sanitization
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			_ = rc.Close()
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			_ = out.Close()
			_ = rc.Close()
			return err
		}
		_ = out.Close()
		_ = rc.Close()
	}
	return nil
}

func main() {
	const extractDir = "extract"
	const indexHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>ZipSlip | CyberPoC</title>
</head>
<body>
<h1>ZipSlip Challenge</h1>
<p>上传一个 ZIP 文件，服务器会将其解压到 extract/ 目录。</p>
<form action="/unzip" method="post" enctype="multipart/form-data">
    <input type="file" name="file">
    <button type="submit">Unzip</button>
    <p>提示：服务器在解压后会执行 <code>./extracted.sh</code> 并显示输出。</p>
</form>
</body>
</html>`

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, indexHtml)
	})

	http.HandleFunc("/unzip", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte("Method Not Allowed"))
			return
		}
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("bad form"))
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("missing file"))
			return
		}
		defer file.Close()

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, file); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("read error"))
			return
		}

		if err := unzipUnsafe(extractDir, buf.Bytes()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(w, "unzip error: %v", err)
			return
		}

		cmd := exec.Command("sh", "extracted.sh")
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(w, "post script err: %s, stderr: %s", err.Error(), stderr.String())
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, `<h2>Unzipped!</h2><br/>
<pre>%s</pre>
<a href='/' >Home</a>
`, out.String())
	})

	if _, err := os.Stat(extractDir); err != nil {
		if err := os.MkdirAll(extractDir, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	flag := os.Getenv("flag")
	if err := writeFile("flag", []byte(flag), 0644); err != nil {
		log.Fatal(err)
	}
	if err := writeFile("extracted.sh", []byte("echo extracted"), 0755); err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":80", nil))
}

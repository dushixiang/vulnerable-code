package main

import (
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
)

func main() {
	var flag = os.Getenv("flag")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "Guest"
		}

		tmpl, err := template.New("greet").Parse("<h1>Hello, " + name + "</h1><p>欢迎访问我们的网站，现在的时间是: {{.Now}}</p>")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]string{
			"Now":  time.Now().Format("2006-01-02 15:04:05"),
			"Flag": flag,
		}
		_ = tmpl.Execute(w, data)
	})

	log.Println("Server starting on :80")
	if err := http.ListenAndServe(":80", nil); err != nil {
		log.Fatal(err)
	}
}

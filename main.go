package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func init() {
	// create initial needed folders
	if err := os.MkdirAll("data/files", 0700); err != nil {
		panic(err)
	}
	if err := os.MkdirAll("data/attachments", 0700); err != nil {
		panic(err)
	}
}

func main() {
	srv := http.Server{
		Addr:         ":8000",
		Handler:      notes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func notes() http.HandlerFunc {
	// read index.html at start
	// so it doesn't have to be read on every page load
	index, err := ioutil.ReadFile("web/index.html")
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// index
		if path == "/" || path == "/index.html" {
			// TODO: Show all files
			// https://gist.github.com/mxlje/8e6279a90dc8f79f65fa8c855e1d7a79
			http.Error(w, "usage: /<name>.md", http.StatusBadRequest)
			return
		}

		// markdown files
		if strings.HasSuffix(path, ".md") {
			// all files will be lowercase
			path = "data/files" + strings.ToLower(path)

			switch r.Method {
			case http.MethodGet:
				file, err := os.Open(path)
				if err != nil {
					if !os.IsNotExist(err) {
						http.Error(w, err.Error(),
							http.StatusInternalServerError)
						return
					}

					parts := strings.Split(path, "/")
					err = os.MkdirAll(
						strings.Join(parts[:len(parts)-1], "/"), 0700)
					if err != nil {
						http.Error(w, err.Error(),
							http.StatusInternalServerError)
						return
					}

					file, err = os.Create(path)
					if err != nil {
						http.Error(w, err.Error(),
							http.StatusInternalServerError)
						return
					}
				}

				fileBytes, err := ioutil.ReadAll(file)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Write(bytes.ReplaceAll(index, []byte("{{VALUE}}"), fileBytes))
			case http.MethodPut:
				file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0700)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				_, err = io.Copy(file, r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			return
		}

		// attachments
		switch r.Method {
		case http.MethodGet:
			// if it starts with /attachments/ it's an attachment
			// and we need to serve it from the attachment folder
			if strings.HasPrefix(path, "/attachments/") {
				parts := strings.Split(path, "/")
				name := parts[len(parts)-1]

				http.ServeFile(w, r, "data/attachments/"+name)
				return
			}

			// serve static assets from /web folder
			http.ServeFile(w, r, "web"+path)
		case http.MethodPost:
			if path != "/attachments" {
				http.Error(w, "can only POST to /attachments",
					http.StatusBadRequest)
				return
			}

			err := r.ParseMultipartForm(10_000_000)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// in your case file would be fileupload
			file, header, err := r.FormFile("file")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// use it as a unique specifier for attachments, otherwise
			// files with the same name could be overridden
			id := strconv.Itoa(int(time.Now().UnixNano()))
			name := "/attachments/" + id + "_" + header.Filename
			localFile, err := os.OpenFile("data"+name,
				os.O_WRONLY|os.O_CREATE, 0700)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = io.Copy(localFile, file)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(map[string]string{
				"file": name,
			})
		}
	}
}

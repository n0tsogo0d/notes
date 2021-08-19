package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	srv := http.Server{
		Addr:         ":8000",
		Handler:      writer(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func writer() http.HandlerFunc {
	index, err := ioutil.ReadFile("dist/index.html")
	if err != nil {
		panic(err)
	}
	indexParts := bytes.Split(index, []byte("{{VALUE}}"))

	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/" {
			// https://gist.github.com/mxlje/8e6279a90dc8f79f65fa8c855e1d7a79
			// render recursive tree
			http.Error(w, "usage: /<name>.md", http.StatusBadRequest)
			return
		}

		if strings.HasSuffix(path, ".md") {
			// all files will be lowercase
			path = "data/files" + strings.ToLower(path)

			switch r.Method {
			case http.MethodGet:
				file, err := os.Open(path)
				if err != nil {
					if os.IsNotExist(err) {
						parts := strings.Split(path, "/")

						// create all folders
						err = os.MkdirAll(strings.Join(
							parts[:len(parts)-1], "/"), 0700)
						if err != nil {
							http.Error(w, err.Error(),
								http.StatusInternalServerError)
							return
						}
						// investigate
						file, err = os.Create(path)
						if err != nil {
							http.Error(w, err.Error(),
								http.StatusInternalServerError)
							return
						}
						return
					}

					http.Error(w, err.Error(), http.StatusInternalServerError)
				}

				bytes, err := ioutil.ReadAll(file)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				// this looks ugly
				// explanation: indexPart[0] + bytes + indexParts[1]
				w.Write(append(indexParts[0],
					append(bytes, indexParts[1]...)...))
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
			default:
				http.Error(w, fmt.Sprintf("unsupported method: %s", r.Method),
					http.StatusInternalServerError)
			}
			return
		}

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

			// serve default file
			http.ServeFile(w, r, "dist"+path)
		case http.MethodPost:
			if path != "/attachments" {
				http.Error(w, "can only POST to /attachments",
					http.StatusBadRequest)
				return
			}

			// this will be the id of the asset
			id := strconv.Itoa(int(time.Now().UnixNano()))

			err := r.ParseMultipartForm(10_000_000) // limit your max input length!
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// in your case file would be fileupload
			file, _, err := r.FormFile("file")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			loc := fmt.Sprintf("data%s/%s", path, id)
			localFile, err := os.OpenFile(loc, os.O_WRONLY|os.O_CREATE, 0700)
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
				"file": fmt.Sprintf("/attachments/%s", id),
			})
		default:
			http.Error(w, fmt.Sprintf("unsupported method: %s", r.Method),
				http.StatusInternalServerError)
		}
	}
}

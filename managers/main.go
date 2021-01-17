package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

func main() {
	h1 := func(w http.ResponseWriter, req *http.Request) {
		mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		if err != nil {
			log.Fatal(err)
		}
		if strings.HasPrefix(mediaType, "multipart/") {
			field := params["boundary"]
			mr := multipart.NewReader(req.Body, field)
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					return
				}
				if err != nil {
					log.Fatal(err)
				}
				slurp, err := ioutil.ReadAll(p)
				if err != nil {
					log.Fatal(err)
				}

				_, pparams, err := mime.ParseMediaType(p.Header.Get("Content-Disposition"))

				if err != nil {
					log.Fatal(err)
				}
				if value, ok := pparams["filename"]; ok {
					fmt.Println("name, filename:", pparams["name"], ", ", value)
					fmt.Printf("slurp: %s", slurp)
					writeFile("./"+value+".copy2", slurp)
				}
			}
		}
	}

	http.HandleFunc("/", h1)

	log.Fatal(http.ListenAndServe(":8090", nil))
}

func writeFile(n string, d []byte) {
	fd, err := os.Create(n)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(fd, bytes.NewReader(d))
	if err != nil {
		panic(err)
	}

	fd.Close()
}

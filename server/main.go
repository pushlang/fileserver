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
					writeFile(value+".copy", slurp)

					client := &http.Client{}

					//prepare the reader instances to encode
					values := map[string]io.Reader{
						pparams["name"]: mustOpen(value+".copy"), // lets assume its this file
						"other":         strings.NewReader("some information"),
					}
					err := Upload(client, "http://127.0.0.1:8090", values)
					if err != nil {
						panic(err)
					}
				}
			}

		}
	}

	http.HandleFunc("/", h1)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Upload(client *http.Client, url string, values map[string]io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
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

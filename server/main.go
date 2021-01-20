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
	"path/filepath"
	"strings"
)

func mustParseMT(r *http.Request, h string) (string, map[string]string) {
	t, p, err := mime.ParseMediaType(r.Header.Get(h))
	if err != nil {
		panic(err)
	}
	return t, p
}

func pathExec(fn string, tl string) string {
	fmt.Println("filename:", fileName)
	p, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(path) + "/" + fn + tl
}

func main() {
	h1 := func(w http.ResponseWriter, req *http.Request) {
		mtype, params := mustParseMT(req, "Content-Type")
		if strings.HasPrefix(mtype, "multipart/") {
			partId := params["boundary"]
			mr := multipart.NewReader(req.Body, partId)
			for {
				part, err := mr.NextPart()
				if err == io.EOF {
					return
				}
				if err != nil {
					log.Fatal(err)
				}
				partData, err := ioutil.ReadAll(part)
				if err != nil {
					log.Fatal(err)
				}
				_, params = mustParseMT(part, "Content-Disposition")

				if fileName, ok := params["filename"]; ok {
					fullPath := pathExec(fileName, ".copy")
					fmt.Println("fullPath:", fullPath)
					writeFile(fullPath, partData)

					//prepare the reader instances to encode
					values := map[string]io.Reader{
						pparams["name"]: mustOpen(fullPath),
						"username":      strings.NewReader("user01"),
					}
					err = Upload(&http.Client{},
						"http://127.0.0.1:8090", values)
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
				return err
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return err
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
		return err
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return err
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

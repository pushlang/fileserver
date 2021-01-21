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

var manager = "http://127.0.0.1:8090"

func mustParseMT(r *http.Request, h string) (string, map[string]string) {
	t, p, err := mime.ParseMediaType(r.Header.Get(h))
	if err != nil {
		panic(err)
	}
	return t, p
}

func pathExec(fn string, tl string) string {
	fmt.Println("file name:", fn)
	p, err := os.Executable()
	if err != nil {
		panic(err)
	}
	fp := filepath.Dir(p) + "/" + fn + tl

	fmt.Println("full path:", fp)
	return fp
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
					writeFile(fullPath, partData)

					values := map[string]io.Reader{
						params["name"]: mustOpen(fullPath),
						"username":     strings.NewReader("user01"),
					}
					err = Upload(&http.Client{}, manager, values)

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
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return
		}

	}
	w.Close()
	postForm(url, b, w)
}
func postForm(url string, b bytes.Buffer, w multipart.NewWriter) (err error) {
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		return
	}

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

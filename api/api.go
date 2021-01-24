package api

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

func HandlerServer(w http.ResponseWriter, req *http.Request) {
	mtype, params := mustParseMT(req.Header.Get("Content-Type"))
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
			_, params = mustParseMT(part.Header.Get("Content-Disposition"))

			if fileName, ok := params["filename"]; ok {
				fullPath := pathExec(fileName, ".copy")
				writeFile(fullPath, partData)

				values := map[string]io.Reader{
					params["name"]: MustOpen(fullPath),
					"username":     strings.NewReader("user01"),
				}
				err = UploadForm(manager, values)

				if err != nil {
					panic(err)
				}
			}
		}

	}
}
func HandlerManager(w http.ResponseWriter, req *http.Request) {
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

				path, err := os.Executable()
				fmt.Println("path:", path)

				if err != nil {
					log.Fatal(err)
				}

				value2 := filepath.Dir(path) + "/" + filepath.Base(value) + "2"
				fmt.Println("value2:", value2)
				writeFile(value2, slurp)
			}
		}
	}
}

func UploadForm(url string, values map[string]io.Reader) (err error) {
	b, w, err := createForm(values)
	err = postForm(url, b, w)
	return
}

func createForm(values map[string]io.Reader) (bytes.Buffer, *multipart.Writer, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	var err error

	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return bytes.Buffer{}, nil, err
			}
		} else {
			if fw, err = w.CreateFormField(key); err != nil {
				return bytes.Buffer{}, nil, err
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return bytes.Buffer{}, nil, err
		}
	}
	w.Close()
	return b, w, nil
}

func postForm(url string, b bytes.Buffer, w *multipart.Writer) (err error) {
	c := &http.Client{}
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := c.Do(req)
	if err != nil {
		return
	}
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

func mustParseMT(h string) (string, map[string]string) {
	t, p, err := mime.ParseMediaType(h)
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

func MustOpen(f string) *os.File {
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

package api

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
)

//var mgr = "http://127.0.0.1:8090"
var mgr = ":8090"

func HandlerServer(w http.ResponseWriter, req *http.Request) {
	mtype, params := mustParseMT(req.Header.Get("Content-Type"))
	if strings.HasPrefix(mtype, "multipart/") {
		partId := params["boundary"]
		mr := multipart.NewReader(req.Body, partId)
		values := map[string]io.Reader{}
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				return
			}
			if err != nil {
				panic(err)
			}
			partData, err := ioutil.ReadAll(part)
			if err != nil {
				panic(err)
			}
			_, params = mustParseMT(part.Header.Get("Content-Disposition"))

			switch params["name"] {
			case "file":
				fullPath := writeFile(params["file"], partData)
				values["file"] = MustOpen(fullPath)
			case "username":
				values["username"] = strings.NewReader(params["username"])
			default:
				panic("Header parameter hasn't recognized")
			}
		}
		if err := UploadForm(mgr, values); err != nil {
			panic(err)
		}
	}
}

func HandlerManager(w http.ResponseWriter, req *http.Request) {
	mtype, params := mustParseMT(req.Header.Get("Content-Type"))
	if strings.HasPrefix(mtype, "multipart/") {
		partId := params["boundary"]
		mr := multipart.NewReader(req.Body, partId)
		values := map[string]io.Reader{}
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				return
			}
			if err != nil {
				panic(err)
			}
			partData, err := ioutil.ReadAll(part)
			if err != nil {
				panic(err)
			}
			_, params = mustParseMT(part.Header.Get("Content-Disposition"))

			switch params["name"] {
			case "file":
				fullPath := writeFile(params["file"], partData)
				values["file"] = MustOpen(fullPath)
			case "username":
				values["username"] = strings.NewReader(params["username"])
			default:
				panic("Header parameter hasn't recognized")
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

	printRequest(req)

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

func MustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}

func writeFile(n string, d []byte) string {
	fmt.Println("file name (writefile):", n)
	fp := pathExec(n)
	fd, err := os.Create(fp)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(fd, bytes.NewReader(d))
	if err != nil {
		panic(err)
	}

	fd.Close()
	return fp
}

func pathExec(n string) string {
	fmt.Println("file name:", n)
	p, err := os.Executable()
	if err != nil {
		panic(err)
	}
	fp := filepath.Dir(p) + "/" + n

	fmt.Println("full path:", fp)
	return fp
}

func printRequest(r *http.Request) {
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", b)
}

package fileserver

import (
	"fileserver/api"
	"io"
	"strings"
)

var host = "http://127.0.0.1:8080"

func RunClient(filename string, username string) {
	values := map[string]io.Reader{
		"file":     api.MustOpen(filename),
		"username": strings.NewReader(username),
	}
	err := api.UploadForm(host, values)
	if err != nil {
		panic(err)
	}
}

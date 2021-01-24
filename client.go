package fileserver

import (
	"fileserver/api"
	"io"
	"strings"
)

func RunClient() {
	values := map[string]io.Reader{
		"file":       api.MustOpen("main.go"),
		"commentary": strings.NewReader("it's main.go"),
	}
	err := api.UploadForm(srv, values)
	if err != nil {
		panic(err)
	}
}

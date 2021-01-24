package fileserver

import (
	"fileserver/api"
	"log"
	"net/http"
)

var mgr = ":8090"

func RunManager() {
	http.HandleFunc("/", api.HandlerManager)
	log.Fatal(http.ListenAndServe(mgr, nil))
}

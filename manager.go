package fileserver

import (
	"fileserver/api"
	"log"
	"net/http"
)

//var mgr = "http://127.0.0.1:8090"
var mgr = ":8090"

func RunManager() {
	http.HandleFunc("/", api.HandlerManager)
	log.Fatal(http.ListenAndServe(mgr, nil))
}

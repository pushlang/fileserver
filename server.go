package fileserver

import (
	"fileserver/api"
	"log"
	"net/http"
)

//var srv = "http://127.0.0.1:8080"
var httpsrv = "127.0.0.1:8080"

func RunServer() {
	http.HandleFunc("/", api.HandlerServer)
	log.Fatal(http.ListenAndServe(httpsrv, nil))
}

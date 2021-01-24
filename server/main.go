package main

import (
	"fileserver"
	"fmt"
)

func main() {
	fmt.Println("Server")
	fileserver.RunServer()
	//fileserver.RunManager()
	//fileserver.RunClient("main.go", "user01")
}

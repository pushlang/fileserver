package main

import (
	"fileserver"
	"fmt"
)

func main() {
	fmt.Println("Manager")
	//fileserver.RunServer()
	fileserver.RunManager()
	//fileserver.RunClient("main.go", "user01")
}

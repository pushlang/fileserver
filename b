#!/bin/bash
go build -o ./client/c ./client/main.go 
go build -o ./manager/m ./manager/main.go
go build -o ./server/s ./server/main.go

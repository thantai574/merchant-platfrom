#!/usr/bin/env bash

go get github.com/githubnemo/CompileDaemon
pwd

go mod download
go mod tidy
go mod vendor


dt=$(date '+%d/%m/%Y %H:%M:%S')
echo "Run dev $dt"

CompileDaemon -log-prefix=false -build="go build -i -x cmd/api/main.go" -command="./main" -exclude-dir=".git"  -exclude-dir=".idea" -exclude-dir="vendor" -color

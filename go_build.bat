#! zsh

echo "Building linux"

go build

echo "Building Windows"
GOOS=windows GOARCH=amd64 go build


VERSION=0.1
GOOS=darwin GOARCH=amd64 go build
mv screencastinator screencastinator-v$VERSION-darwin-amd64
GOOS=linux GOARCH=386 go build
mv screencastinator screencastinator-v$VERSION-linux-386
GOOS=linux GOARCH=amd64 go build
mv screencastinator screencastinator-v$VERSION-linux-amd64


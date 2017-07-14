GOOS=linux GOARCH=amd64 go build -ldflags "-X main.buildstamp=$(date +%FT%T%z) -X main.githash=$(git rev-parse HEAD)"

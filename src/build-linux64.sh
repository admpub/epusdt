export GOOS=linux
export GOARCH=amd64
go build -o ./runtime/epusdt_${GOOS}_${GOARCH} --trimpath --ldflags="-w -s"
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build  -o release/win64/biu.exe
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build  -o release/win32/biu.exe
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build  -o release/darwin64/biu
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o release/linux64/biu
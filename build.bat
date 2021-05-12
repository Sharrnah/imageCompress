set GOARCH=amd64
set GOOS=windows
set CGO_ENABLED=1
set GOTMPDIR=%cd%

go build -tags osusergo -v -ldflags "-s -w" -o bin/imageResize.exe

echo go build -tags osusergo -v -compiler=gccgo -gccgoflags "-s -w" -o bin/imageResize_gccgo.exe

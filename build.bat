set GOARCH=amd64
set GOOS=windows
set CGO_ENABLED=1
set GOTMPDIR=%cd%

set origdir=%cd%

cd bin
windres -o imageCompress-res.syso imageCompress.rc
cd %origdir%
go build -tags osusergo -v -ldflags "-s -w" -o bin/imageCompress.exe

echo go build -tags osusergo -v -compiler=gccgo -gccgoflags "-s -w" -o bin/imageCompress_gccgo.exe

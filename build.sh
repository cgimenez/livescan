echo "Building for Windows"
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ HOST=x86_64-w64-mingw32 go build -ldflags "-extldflags=-static" -p 4 -v -o bin/livescan.exe

echo "Building for Mac OS"
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -a -o bin/livescan-osx
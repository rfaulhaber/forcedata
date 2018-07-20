GOBUILD = go build -v 
SOURCE = github.com/rfaulhaber/force-data
ARCH = arch=amd64

all: linux mac windows freebsd

linux: main.go
	env GOOS=linux $(ARGH) $(GOBUILD) -o out/data-linux $(SOURCE)

mac: main.go
	env GOOS=darwin $(ARCH) $(GOBUILD) -o out/data-mac $(SOURCE)

windows: main.go
	env GOOS=windows $(ARCH) $(GOBUILD) -o out/data-windows.exe $(SOURCE)

freebsd: main.go
	env GOOS=freebsd $(ARCH) $(GOBUILD) -o out/data-freebsd $(SOURCE)

.PHONY: clean

clean:
	rm -rf ./out

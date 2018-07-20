GOBUILD = go build -v 
SOURCE = github.com/rfaulhaber/force-data
ARCH = arch=amd64
OUTDIR = ./out

all: linux mac windows freebsd

linux: main.go
	env GOOS=linux $(ARGH) $(GOBUILD) -o $(OUTDIR)/data-linux $(SOURCE)

mac: main.go
	env GOOS=darwin $(ARCH) $(GOBUILD) -o $(OUTDIR)/data-mac $(SOURCE)

windows: main.go
	env GOOS=windows $(ARCH) $(GOBUILD) -o $(OUTDIR)/data-windows.exe $(SOURCE)

freebsd: main.go
	env GOOS=freebsd $(ARCH) $(GOBUILD) -o $(OUTDIR)/data-freebsd $(SOURCE)

.PHONY: clean

clean:
	rm -rf $(OUTDIR)

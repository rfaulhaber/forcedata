GOBUILD = go build -v
SOURCE = github.com/rfaulhaber/forcedata
ARCH = arch=amd64
OUTDIR = ./out
CMDNAME = forcedata

all: linux mac windows freebsd

linux: main.go
	env GOOS=linux $(ARGH) $(GOBUILD) -o $(OUTDIR)/$(CMDNAME)-linux

mac: main.go
	env GOOS=darwin $(ARCH) $(GOBUILD) -o $(OUTDIR)/$(CMDNAME)-mac

windows: main.go
	env GOOS=windows $(ARCH) $(GOBUILD) -o $(OUTDIR)/$(CMDNAME)-windows.exe

freebsd: main.go
	env GOOS=freebsd $(ARCH) $(GOBUILD) -o $(OUTDIR)/$(CMDNAME)-freebsd

.PHONY: clean

clean:
	rm -rf $(OUTDIR)

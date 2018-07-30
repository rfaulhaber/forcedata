# forcedata
[![pipeline status](https://gitlab.com/rfaulhaber/forcedata/badges/master/pipeline.svg)](https://gitlab.com/rfaulhaber/forcedata/commits/master)

CLI tool for manipulating data in Salesforce

## Install
Download one of the release binaries and add it to a directory in your PATH.

## Usage

### Obtaining REST credentials

This program uses Salesforce's Bulk API, which requires credentials.

## Building 
Assuming you have a 
[properly configured Go environment](https://golang.org/doc/code.html), run:

```
go get -u github.com/rfaulhaber/forcedata
```

You may now either run `go build` or `go install` to build a binary or install 
a binary to your GOBIN, respectively.

## Roadmap
This is a list of all the things I'd like to complete before I consider this to 
be v1.0, subject to change:

- [ ] Implement authentication prompts
- [ ] Implement finished job reporting (GetSuccess, GetFailure, GetUnprocessed)
 via `report` command.
- [ ] Allow for multiple files to be specified in `load`
- [ ] Release the `job` package as a separate repo
- [ ] Write more comprehensive documentation for `auth` and `job` packages.
- [ ] Generate bash / zsh completions 
- [ ] Write GoDoc
- [ ] Write install script

## License
This program is freely distributed under the MIT license with the hope that it 
will be useful.

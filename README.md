# forcedata
[![pipeline status](https://gitlab.com/rfaulhaber/forcedata/badges/master/pipeline.svg)](https://gitlab.com/rfaulhaber/forcedata/commits/master)
[![coverage report](https://gitlab.com/rfaulhaber/forcedata/badges/master/coverage.svg)](https://gitlab.com/rfaulhaber/forcedata/commits/master)

CLI tool for loading data in Salesforce. Currently under development.

This was written mostly as an exercise in writing Go. The Salesforce CLI
does everything this can do right now.

## Install
forcedata should work on [any OS that Go builds for](https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63).
As of now it doesn't have any special OS dependencies.

I don't have an install script yet so you will have to install a binary 
manually. Download one of the release binaries and add it to a directory in 
your PATH. On a *nix system, you can do this by doing something like:

```
mv forcedata-linux /usr/local/bin/data
```

## Usage
You are encouraged to run `data --help` for the help text for the program overall. You may specify the `--help` flag to 
see the help text for every subsequent command (e.g. `data load --help`).

The following commands are available: 

- `authenticate` - for generating an oauth access token (see below)
- `load` - for creating Bulk API jobs
- `version` - prints the current version and exits

### Obtaining REST credentials

This program uses Salesforce's Bulk API, which requires oauth authentication.

First, you must create a [connected App in Salesforce](https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro_defining_remote_access_applications.htm). 
You'll then need a JSON file with the following fields:`client_id`, `client_secret`, `username`, and `password`,
corresponding to the client ID and secret of your connected app, and the username and password of the user you want to
run the program as.

Once you have a JSON file, you can now run:

```
data authenticate --file your-creds.json
``` 
This will generate an access
token. The program will write a JSON file containing an access token and some other information to stdout. You will need
this JSON file to run any other commands.

This token will eventually expire, once that happens you can rerun the previous command to get a new one. 

## Building 
Assuming you have a [properly configured Go environment](https://golang.org/doc/code.html), run:

```
go get -u github.com/rfaulhaber/forcedata
```

You may now either run `go build` or `go install` to build a binary or install 
a binary to your GOBIN, respectively.

This repo also includes a makefile for convenient compiling to common operating
systems.

## Roadmap
This is a list of all the things I'd like to complete before I consider this to 
be v1.0, subject to change:

- [ ] Implement authentication prompts
- [ ] Implement finished job reporting (GetSuccess, GetFailure, GetUnprocessed)
 via `report` command
- [ ] Allow for multiple files to be specified in `load`
- [ ] Release the `job` package as a separate repo
- [ ] Write more comprehensive documentation for `auth` and `job` packages
- [ ] Write better tests
- [ ] Generate bash / zsh completions 
- [ ] Write GoDoc
- [ ] Write install script

## License
This program is freely distributed under the MIT license with the hope that it 
will be useful. This program is not endorsed or affiliated with Salesforce.

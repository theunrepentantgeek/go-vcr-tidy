# go-vcr-tidy

Extensions for the popular library [go-vcr](https://github.com/dnaeon/go-vcr) to reduce the size of cassettes. 

The goal is to improve test replay times by providing a set of tools that can be used to reduce the size of recordings by eliding specific interactions from the recording. The initial focus will be on tools that work with Azure interactions (because that's my need), but pull requests for other patterns will be welcomed.

## Project Status

Proof-of-concept status, still working on testing the effects of cassette reduction.

## Evaluation

The `go-vcr-tidy` CLI provides a way to apply selected reductions to existing cassette files, allowing you to test the effectiveness of `go-vcr-tidy` for your testing needs.

Install `go-vcr-tidy` from source:

``` bash
go install github.com/theunrepentantgeek/go-vcr-tidy@latest
```

``` bash
$ go-vcr-tidy clean --help
Usage: go-vcr-tidy clean <globs> ... [flags]

Clean go-vcr cassette files.

Arguments:
  <globs> ...    Paths to go-vcr cassette files to clean. Globbing allowed.

Flags:
  -h, --help                             Show context-sensitive help.
  -v, --verbose                          Enable verbose logging.

      --clean.deletes                    Clean delete interactions.
      --clean.long-running-operations    Clean Azure long-running operation interactions.
      --clean.resource-modifications     Clean Azure resource modification (PUT/PATCH) monitoring interactions.
      --clean.resource-deletions         Clean Azure resource deletion monitoring interactions.
```

## Quick Start

Add the `go-vcr-tidy` package to your project

``` bash
go get github.com/theunrepentantgeek/go-vcr-tidy
```

When creating your `go-vcr` `recorder`, create create a `Cleaner` and hook that into your `recorder`:

``` go
TBC
```

## Supported Tools

### Monitoring for deletion

After a client issuing a DELETE, a client issues repeated GET requests to the same URL, waiting for a final GET to return a 404 indicating deletion of the resource has completed. 

**Status**: Implemented, pending testing

### Post creation and upate monitoring in Azure

After issuing a PUT to an Azure service, a client issues repeated GET requests, waiting for the `provisioningState` of the resoruce to change from `Creating` to something else (hopefully `Succeeded`).

**Status**: Implemented, pending testing

## Azure long running operations

If a PUT to an Azure service returns a long-running-operation (LRO), the client will poll that operation until it has completed.  We can shorten the runtime of the LRO - effectively turning it into a short-running-operation (SRO) - by eliding most of the waiting time.

**Status**: Implemented, pending testing

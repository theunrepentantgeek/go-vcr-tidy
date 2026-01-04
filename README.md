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
$ go-vcr-tidy --help
Usage: go-vcr-tidy <globs> ... [flags]

Arguments:
  <globs> ...    Paths to go-vcr cassette files to clean. Globbing allowed.

Flags:
  -h, --help                                   Show context-sensitive help.
  -v, --verbose                                Enable verbose logging.
      --clean.deletes                          Clean delete interactions.
      --clean.azure.all                        Clean all Azure-related monitoring interactions.
      --clean.azure.long-running-operations    Clean Azure long-running operation interactions.
      --clean.azure.resource-modifications     Clean Azure resource modification (PUT/PATCH) monitoring interactions.
      --clean.azure.resource-deletions         Clean Azure resource deletion monitoring interactions.
```

On Linux and MacOS, globs should be "double quoted" to prevent the shell from expanding the wildcard.

## Quick Start

Add the `go-vcr-tidy` package to your project

``` bash
go get github.com/theunrepentantgeek/go-vcr-tidy
```

When creating your `recorder`, create a `Cleaner` with appropriate options and hook that into your `recorder`:

``` go
TBC
```

## Options

A number of different cleaning strategies are available - select the ones that make sense for your context.

### Monitored deletes

This analyzer looks for the following pattern of HTTP interactions:

| Stage   |    HTTP Method     | Status | Note            |
| ------- | :----------------: | :----: | --------------- |
| Trigger | DELETE &lt;url&gt; |  2xx   |                 |
| Monitor |  GET &lt;url&gt;   |  2xx   | Repeats n times |
| Finish  |  GET &lt;url&gt;   |  404   |                 |

When detected, the first and last GET 2xx are retained, with the intervening ones removed.

### Post creation and update monitoring in Azure

This analyzer looks for the following pattern of HTTP interactions:

| Stage   |                HTTP Method                | Status | Note                                                |
| ------- | :---------------------------------------: | :----: | --------------------------------------------------- |
| Trigger | PUT &lt;url&gt;<br/> or PATCH &lt;url&gt; |  2xx   |                                                     |
| Monitor |              GET &lt;url&gt;              |  2xx   | `provisioningState` is `Creating`, repeated n times |
| Finish  |              GET &lt;url&gt;              |  2xx   | `provisioningState` has changed                     |

When detected, the first and last GET 2xx with `provisioningState` `Creating` are retained, and the intervening ones removed.

## Azure long running operation

This analyzer looks for the following pattern of HTTP interactions:

| Stage   |                             HTTP Method                             | Status | Note                                    |
| ------- | :-----------------------------------------------------------------: | :----: | --------------------------------------- |
| Trigger | PUT &lt;url&gt;<br/>or PATCH &lt;url&gt; <br/>or DELETE &lt;url&gt; |  2xx   | Returning `Azure-Asyncoperation` header |
| Monitor |                        GET &lt;operation&gt;                        |  2xx   | Operation status is `InProgress`        |
| Finish  |                        GET &lt;operation&gt;                        |  2xx   | Operation status has changed            |

When detected, the first and last GET of the operation URL are retained, and the intervening ones removed.

## Azure asynchronous operation

This analyzer looks for the following pattern of HTTP interactions:

| Stage   |                             HTTP Method                              | Status | Note                             |
| ------- | :------------------------------------------------------------------: | :----: | -------------------------------- |
| Trigger | PUT &lt;url&gt;<br/> or PATCH &lt;url&gt; <br/>or DELETE &lt;url&gt; |  2xx   | Returning `Location` header      |
| Monitor |                         GET &lt;location&gt;                         |  2xx   | Operation status is `InProgress` |
| Finish  |                         GET &lt;location&gt;                         |  2xx   | Operation status has changed     |

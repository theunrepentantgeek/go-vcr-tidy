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
      --clean-deletes                          Clean delete interactions.
      --clean-azure-all                        Clean all Azure-related monitoring interactions.
      --clean-azure-long-running-operations    Clean Azure long-running operation interactions.
      --clean-azure-resource-modifications     Clean Azure resource modification (PUT/PATCH) monitoring interactions.
      --clean-azure-resource-deletions         Clean Azure resource deletion monitoring interactions.
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

## Cleaning Strategies

A number of different cleaning strategies are available - select the ones that make sense for your context.

### Monitored deletes

Client issues a DELETE request for a resource, then polls the resource URL with repeated GET requests until the resource is confirmed deleted (404).

| Stage   |    HTTP Method     | Status | Note            |
| ------- | :----------------: | :----: | --------------- |
| Trigger | DELETE &lt;url&gt; |  2xx   |                 |
| Monitor |  GET &lt;url&gt;   |  2xx   | Repeats n times |
| Finish  |  GET &lt;url&gt;   |  404   |                 |

Retain the initial DELETE request, the first and last successful GET requests, and the final 404 GET request. Remove the intervening successful GET requests.

Enable on the CLI with `--clean-deletes` or in code by passing the `ReduceDeleteMonitoring()` option to `vcrcleaner.New()`.

### Post creation and update monitoring in Azure

Client issues a PUT or PATCH request to create or update an Azure resource, then polls the resource URL with repeated GET requests until the resource's `provisioningState` changes from `Creating` to another state.

| Stage   |                HTTP Method                | Status | Note                                                |
| ------- | :---------------------------------------: | :----: | --------------------------------------------------- |
| Trigger | PUT &lt;url&gt;<br/> or PATCH &lt;url&gt; |  2xx   |                                                     |
| Monitor |              GET &lt;url&gt;              |  2xx   | `provisioningState` is `Creating`, repeated n times |
| Finish  |              GET &lt;url&gt;              |  2xx   | `provisioningState` has changed                     |

Retain the initial PUT/PATCH request, the first and last successful GET requests with `provisioningState` `Creating`, and the final GET request where `provisioningState` has changed. Remove the intervening successful GET requests with `provisioningState` `Creating`.

Enable on the CLI with `--clean-azure-resource-modifications` or in code by passing the `ReduceAzureResourceModificationMonitoring()` option to `vcrcleaner.New()`.

### Post deletion monitoring in Azure

Client issues a DELETE request to delete an Azure resource, then polls the resource URL with repeated GET requests returning a `provisioningState` of `Deleting` until the resource is confirmed deleted (404).

| Stage   |    HTTP Method     | Status | Note                                                |
| ------- | :----------------: | :----: | --------------------------------------------------- |
| Trigger | DELETE &lt;url&gt; |  2xx   |                                                     |
| Monitor |  GET &lt;url&gt;   |  2xx   | `provisioningState` is `Deleting`, repeated n times |
| Finish  |  GET &lt;url&gt;   |  404   |                                                     |

Retain the initial DELETE request, the first and last successful GET requests with `provisioningState` `Deleting`, and the final 404 GET request. Remove the intervening successful GET requests with `provisioningState` `Deleting`.

Enable on the CLI with `--clean-azure-resource-deletions` or in code by passing the `ReduceAzureResourceDeletionMonitoring()` option to `vcrcleaner.New()`.

### Azure long running operation

Client issues a PUT, PATCH or DELETE request to create, update or delete an Azure resource and receives an `Azure-AsyncOperation` header in the response. The client then polls the operation URL with repeated GET requests until the operation status changes from `InProgress` to another state.

| Stage   |                             HTTP Method                             | Status | Note                                     |
| ------- | :-----------------------------------------------------------------: | :----: | ---------------------------------------- |
| Trigger | PUT &lt;url&gt;<br/>or PATCH &lt;url&gt; <br/>or DELETE &lt;url&gt; |  2xx   | Returning `Azure-AsyncOperation` header |
| Monitor |                        GET &lt;operation&gt;                        |  2xx   | Operation status is `InProgress`        |
| Finish  |                        GET &lt;operation&gt;                        |  2xx   | Operation status has changed            |

Retain the initial PUT/PATCH/DELETE request, the first and last GET requests of the operation with status `InProgress` and remove the intervening ones.

Enable on the CLI with `--clean-azure-long-running-operations` or in code by passing the `ReduceAzureLongRunningOperations()` option to `vcrcleaner.New()`.

### Azure asynchronous operation

Client issues a PUT, PATCH or DELETE request to create, update or delete an Azure resource and receives a `Location` header in the response. The client then polls the location URL with repeated GET requests until the operation status changes from `InProgress` to another state.

| Stage   |                             HTTP Method                              | Status | Note                             |
| ------- | :------------------------------------------------------------------: | :----: | -------------------------------- |
| Trigger | PUT &lt;url&gt;<br/> or PATCH &lt;url&gt; <br/>or DELETE &lt;url&gt; |  2xx   | Returning `Location` header      |
| Monitor |                         GET &lt;location&gt;                         |  2xx   | Operation status is `InProgress` |
| Finish  |                         GET &lt;location&gt;                         |  2xx   | Operation status has changed     |

Retain the initial PUT/PATCH/DELETE request, the first and last GET requests of the location with status `InProgress` and remove the intervening ones.

Enable on the CLI with `--clean-azure-asynchronous-operations` or in code by passing the `ReduceAzureAsynchronousOperationMonitoring()` option to `vcrcleaner.New()`.
# go-vcr-tidy

Extensions for go-vcr to reduce the size of cassettes

The popular library [go-vcr](https://github.com/dnaeon/go-vcr) enables reliable testing by recording the HTTP interactions of the system-under-test to a cassette file which can be replayed on demand.

For large and/or complex systems, these recordings are often large, resulting in tests that can take a long time to run.

This library aims to improve replay times by providing a set of tools that can be used to reduce the size of recordings by eliding selected interactions from the recording. The initial focus will be on tools that work with Azure interactions (because that's my need), but pull requests for other patterns will be welcomed.

Planned tools include:

## Monitoring for creation in Azure

After issuing a PUT to an Azure service, a client might GET repeatedly, waiting for its `provisioningState` to change from `Creating` to something else (hopefully `Succeeded`).

## Monitoring for deletion

After a client issues a DELETE, it might wait for the deletion to complete by issuing a series of GETs to the same URL, waiting for a final GET to return a 404 indicating deletion of the resource has completed. 

We can elide most of those GET requests from the recording, reducing the sequence to DELETE, GET (2xx), Get (404).

## Azure long running operations

If a PUT to an Azure service returns a long-running-operation (LRO), the client will poll that operation until it has completed. 

We can shorten the runtime of the LRO - effectively turning it into a short-running-operation (SRO) - by eliding most of the waiting time.


# projector-hsa-04

Implementation of simple worker that track uah/usd rate and publish to GA via http API;

## Components

**server** - simple web server that server some JS page containing the GA tag. That page was used to generate some
ClientID and SessionID stored under the cookies as GA require ClientID/SessionID to send custom events;

**worker.go** implementation of Rate GA worker;


## How to run?

```bash
docker-compose up
```

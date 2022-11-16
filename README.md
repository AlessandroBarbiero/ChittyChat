# ChittyChat
A simple gRPC implementation of a chat go service

## How to run
Simply run first the server and then how many clients as you want.
For example from command line:

```bash
go run server/server.go
```
```bash
go run client/client.go
```

To make it easier to test we make them work on localhost and port fixed at 8080.

## Output
The output is completely redirected to the log files named `server.log` and `client{$num}.log`.
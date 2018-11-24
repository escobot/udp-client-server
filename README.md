# udp-client-server
http client and file server built on top of udp

## Detailed Usage

### General

#### httpc

``` bash
httpc is a command-line tool built on top of UDP to support HTTP protocol operations.

usage: httpc (get|post) URL [-v] (-h \"k:v\")* [-d inline-data] [-f file] [-udp]

    -v             Prints the detail of the response such as protocol, status, and headers.
    -h key:value   Associates headers to HTTP Request with the format 'key:value'.
    -d string      Associates an inline data to the body HTTP POST request.
    -f file        Associates the content of a file to the body HTTP POST request.
    -o output      Associates the content of a file to the body HTTP POST request.
    -p port        Specifies the port of the server

Either [-d] or [-f] can be used but not both.
```

#### httpfs

``` bash
httpfs is a simple file server built on top of UDP.

usage: httpfs [-v] [-p PORT] [-d PATH-TO-DIR]
    -v Prints debugging messages.
    -p Specifies the port number that the server will listen and serve at. 
    Default is 8080.
    -d Specifies the directory that the server will use to read/write requested files. 
    Default is the current directory when launching the application.
```
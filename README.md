# Gompet - GO Multi-purpose Performance Evaluation Tool

Gompet is a multi-purpose performance evaluation tool which can quickly send 
thousands of commands to servers using reproducibly varying input patterns.

The tool has been optimized for maximum throughput, a single computer can easily 
send tens of thousands of commands/second (with enough network bandwidth).
Optionally the request rate can be limited to N commands/second/client which 
enables more controlled load generation.

The tool reports the latency percentiles for the execution, counts of client
dependent results, as well as counts of different errors received (if any).
The percentiles can also be reported at regular intervals for long running tests.

Gompet currently includes a standard HTTP client, optimized FastHTTP client and 
SQL client for PostgreSQL.

It is easy to add clients for new protocols, and utilize the input variable expansion, 
worker pool management and statistics reporting from the framework. The library 
github.com/jkarjala/gompet can also be imported to applications outside of this repo.

## Installation

Assuming go executable and GOPATH/bin in your PATH:

```
go get -u github.com/jkarjala/gompet/...
```

## General usage

The clients receive commands from command line (each argument is a command, use quotes if 
command has whitespace), or from a file with paramter -f (each line is a command). 
Use "-" as filename for stdin. 

Alternatively, a template with variables $1 - $9 can be given with -t option. In this case, the input 
file/stdin must contain tab-separated-values to be inserted in the variables in the template to 
construct the final command.

See data folder for a few input file examples. For best throughput, use files with more than 500 lines 
(like the number-word examples), otherwise the file open/close overhead skews the results.

The one letter options (and pprof) in usages below are implemented by the framework and thus common
to all clients, the long options are specific to the clients.

### HTTP and FastHTTP Clients

The gompet-http uses the standard Go http library, while the gompet-fasthttp uses the fasthttp 
library. Fasthttp is 20-50% faster but only supports HTTP/1 and reads the whole body in memory,
while the gompet-http reads and discards body unless -v option is given.

```
gompet-http [options] ['cmd 1' 'cmd 2' ...]
  -P    Report progress once a second
  -R int
        Rate limit each client to N queries/sec (accuracy depends on OS)
  -S int
        Show and reset percentiles every N seconds, 0 shows at end
  -auth string
        HTTP Authorization header
  -c int
        Number of parallel clients executing commands (default 1)
  -content-type string
        HTTP body content type (default "application/json")
  -d duration
        Test until given duration elapses, e.g 5m for 5 minutes
  -f string
        Input file name, stdin if '-'
  -pprof
        Enable pprof web server
  -r int
        Repeat the input N times, does not work with stdin (default 1)
  -t string
        Command template, $1-$9 refers to tab-separated columns in input
  -timeout int
        HTTP Client timeout in seconds (default 10)
  -v    Verbose logging

```

The Command syntax for http clients is:

```
HTTP-VERB URL Body-as-single-line-if-needed
```

Examples (after installation)

Run a single HTTP PUT command via command line with geompet-fasthttp and show verbose output:
```
gompet-fasthttp -v 'PUT http://httpbin.org/put { "some" : "put" }'
```

Run HTTP commands from http.txt with verbose output:

```
gompet-http -f testdata/http.txt -v
```

Run HTTP commands with template and the URLs from urls.tsv 
with 2 parallel clients and repeat the file twice: 
```
gompet-http -f testdata/urls.tsv -t 'GET $1' -c 2 -r 2
```

### SQL Client

```
gompet-sql [options] ['cmd 1' 'cmd 2' ...]
  -P    Report progress once a second
  -R int
        Rate limit each client to N queries/sec (accuracy depends on OS)
  -S int
        Show and reset percentiles every N seconds, 0 shows at end
  -c int
        Number of parallel clients executing commands (default 1)
  -d duration
        Test until given duration elapses, e.g 5m for 5 minutes
  -discard
        Discard result set with mimimal memory allocation
  -driver string
        Database driver, 'postgres' or 'mysql'
  -f string
        Input file name, stdin if '-'
  -pprof
        Enable pprof web server
  -r int
        Repeat the input N times, does not work with stdin (default 1)
  -t string
        Command template, $1-$9 refers to tab-separated columns in input
  -tx int
        Batch N commands in one transaction, does not work with SELECTs
  -url string
        SQL Connect URL, e.g. postgres://user:pass@host/db?sslmode=disable
  -v    Verbose logging
```

The Connect URL syntax is 

```
postgres://user:pass@hostname/db?sslmode=disable
```

The Command syntax for SQL client is a single SQL statement. A template with
variables is used as a prepared statement instead of simple text substitution.
Unlike SQL prepared statement, the template may contain multiple references 
to the same variable, they will be duplicated to make the prepared statement 
work.

## Licence

Gompet Copyright 2019 [Jari Karjala](https://www.jarikarjala.com/). 

Gompet is licensed under [GNU General Public License v3](LICENSE).

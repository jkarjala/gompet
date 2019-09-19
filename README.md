# Gompet - Go Multi-purpose Performance Evaluation Tool

Gompet is a multi-core multi-purpose performance evaluation tool which can quickly generate 
heavy load on servers using pre-defined input patterns. It is useful for benchmarking as
well as load testing servers.

Gompet currently includes a standard HTTP client, FastHTTP client and SQL client for PostgreSQL.

It is easy to add new clients and get the worker pool and statistics reporting out of the box. The library github.com/jkarjala/gompet can also be imported to applications outside of this repo.

## Installation

```
go get -u github.com/jkarjala/gompet/...
```

## Usage

The test clients receive commands from command line (each argument is a command), 
or from a file with paramter -f (each line is a command). Use "-" as filename for stdin.

Alternatively, a template with variables $1 - $9 can be given with -t option. In this case, the input 
file/stdin must contain tab-separated-values to be inserted to the variables to construct the command.

See data folder for a few input file examples.

### HTTP and FastHTTP Clients

The gompet-http uses the standard Go http library, while the gompet-fasthttp uses the fasthttp library which is more performant but only supports HTTP/1 and requires valid certificates.

```
Usage of gompet-http and gompet-fasthttp:
  -P    Report progress once a second
  -S    Show and reset percentiles every N seconds, 0 shows at end
  -auth string
        HTTP Authorization header
  -c int
        Number of parallel clients executing commands (default 1)
  -content-type string
        HTTP body content type (default "application/json")
  -f string
        Input file name, stdin if not given or '-'
  -pprof
        enable pprof web server
  -r int
        Repeat the input N times, does not work with stdin (default 1)
  -t string
        Command template, $1-$9 refers to tab-separated columns in input
  -timeout int
        HTTP Client timeout (default 10)
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
Usage of gompet-sql:
  -P    Report progress after every 10k commands
  -c int
        Number of parallel clients executing commands (default 1)
  -discard
        Discard result set with mimimal memory allocation
  -driver string
        Database driver, only 'postgres' supported today
  -f string
        Input file name, stdin if not given or '-' (default "-")
  -pprof
        enable pprof web server
  -t string
        Command template, $1-$9 refers to input file columns (tab-separated)
  -url string
        SQL Connect URL
  -v    Verbose logging
```

The Connect URL syntax is 

```
postgres://user:pass@hostname/db?sslmode=disable
```

The Command syntax for SQL client is a single SQL statement.

## Licence

Gompet Copyright 2019 [Jari Karjala](https://www.jarikarjala.com/). 

Gompet is licensed under [GNU General Public License v3](LICENSE).

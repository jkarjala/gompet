# GOMB - Go Multi-purpose Benchmark

GOMB is a multi-threaded multi-purpose benchmark runner which can quickly generate 
heavy load on servers using pre-defined but varied input patterns. 

GOMB currently includes a standard HTTP client, FastHTTP client and SQL client for PostgreSQL.

It is easy to add new clients and get the worker pool and statistics reporting out of the box.

## Installation

```
go get -u github.com/jkarjala/gomb
```

## Usage

By default the clients read commands from standard input, one command per line. Use -f to read from file instead.

Alternatively, a template with variables $1 - $9 can be given with -t option. In this case, the input must be tab-separated-values to be inserted to the variables to construct the command.

See data folder for a few input file examples.

### HTTP and FastHTTP Clients

```
Usage of gomb-http (or gomb-fasthttp):
  -P    Report progress after every 10k commands
  -auth string
        HTTP Authorization header
  -c int
        Number of parallel clients executing commands (default 1)
  -content-type string
        HTTP body content type (default "application/json")
  -f string
        Input file name, stdin if not given or '-' (default "-")
  -pprof
        enable pprof web server
  -t string
        Command template, $1-$9 refers to input file columns (tab-separated)
  -timeout int
        HTTP Client timeout (default 10)
  -v    Verbose logging
```

The Command syntax is:

```
HTTP-VERB URL Body-as-single-line-if-needed
```

### SQL Client

```
Usage of gomb-sql:
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

The Command syntax is SQL statement.

## Licence

GOMB Copyright 2019 [Jari Karjala](https://www.jarikarjala.com/). 

Licensed under [GNU General Public License v3](LICENSE).

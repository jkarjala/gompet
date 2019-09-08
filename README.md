# GOMB - Go Multi-purpose Benchmark

GOMB is a multi-threaded multi-purpose benchmark runner which can quickly generate 
heavy load on servers using pre-defined but varied input patterns. 

The input commands are either read from file or stdin as-is, or a template with 
variables is given and the variable values are read from file or stdin.

GOMB currently has a generic HTTP client, 
SQL client is under development.

See data folder for a few input file examples.

## Installation

go install bitbucket.com/jkarjala/gomb

## Usage

```
Usage of gomb-http:
  -c int
        Number of parallel clients executing commands (default 1)
  -f string
        Input file name, stdin if not given or '-' (default "-")
  -http-auth string
        HTTP Authorization header
  -http-content-type string
        HTTP POST/PUT content type (default "application/json")
  -t string
        Command template, $1-$9 refers to input file columns (tab-separated)
  -v    Verbose logging
```

The command syntax is:

```
HTTP-VERB   URL     Body-if-needed
```

## Licence

GOMB Copyright 2019 [Jari Karjala](https://www.jarikarjala.com/). 

Licensed under [GNU General Public License v3](LICENSE).

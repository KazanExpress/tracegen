# tracegen [![Actions Status](https://github.com/KazanExpress/tracegen/actions/workflows/go.yaml/badge.svg?branch=master)](https://github.com/KazanExpress/tracegen/actions) [![go report card](https://goreportcard.com/badge/github.com/KazanExpress/tracegen)](https://goreportcard.com/report/github.com/KazanExpress/tracegen) [![PkgGoDev](https://pkg.go.dev/badge/github.com/KazanExpress/tracegen)](https://pkg.go.dev/github.com/KazanExpress/tracegen)

Tool for generating OpenTelemetry tracing decorators.

## Installation

```shell
go get -u github.com/KazanExpress/tracegen/cmd/...
```

## Usage

```
tracegen generates OpenTelemetry tracing decorators.

It scans the public interfaces in the specified source file and generates a 
decorator for each one. Use @trace annotation in the method doc to specify the 
traceable parameters (comma separated list). You can access nested fields and 
types (Array, Bool, Int,Int64, Float64, String, Unknown). If no type is 
specified, Unknown will be used by default. Values with Unknown type will be 
JSON encoded or formatted to string, according to the Go formatting rules, if 
the former is not possible.  Example: "@trace Int64:id". 

Requirements and restrictions:
  - each interface method must have "ctx context.Context" parameter;
  - don't combine parameters of the same type (for example "Do(a, b string)");

Params:
  -destination string
        Output file; defaults to stdout.
  -source string
        Input file; defaults to $GOFILE.
```

## Examples

See [examples](./examples) and [tests](./tracegen/tracegen_test.go) for more information.

## Issue tracker

Please report any bugs and enhancement ideas using the GitHub issue tracker:

https://github.com/KazanExpress/tracegen/issues

Feel free to also ask questions on the tracker.

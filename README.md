# tracegen
Tool for generating OpenTelemetry tracing decorators.

## Installation

```shell
go get -u github.com/KazanExpress/tracegen/cmd/...
```

#### Usage

```
tracegen generates OpenTelemetry tracing decorators.
It scans the public interfaces in the specified source file and generates a 
decorator for each one. Use @trace annotation in the method doc to specify the 
traceable parameters (comma separated list). You can access nested fields and 
types (Array, Bool, Int, Int64, Float64, String, Any). If no type is 
specified, Any will be used by default. Example: "@trace Int64:id". 

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

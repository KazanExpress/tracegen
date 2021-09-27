package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/KazanExpress/tracegen/tracegen"
)

const usageText = `tracegen generates OpenTelemetry tracing decorators.

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
`

func main() {
	source := flag.String("source", "", `Input file; defaults to $GOFILE.`)
	destination := flag.String("destination", "", `Output file; defaults to stdout.`)
	flag.Usage = usage
	flag.Parse()

	if *source == "" {
		*source = os.Getenv("GOFILE")
	}

	output := os.Stdout
	if *destination != "" {
		f, err := os.Create(*destination)
		if err != nil {
			log.Panicf("create destination file: %v", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Panicf("failed to close destination file: %v", err)
			}
		}()
		output = f
	}

	err := tracegen.Run(*source, output)
	if err != nil {
		log.Panicf("run: %v", err)
	}
}

func usage() {
	_, _ = io.WriteString(os.Stderr, usageText)
	flag.PrintDefaults()
}

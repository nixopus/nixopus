# Octopack

Octopack is a tool and library that converts JSON specifications into Dockerfiles. It provides a structured way to define Dockerfile configurations using JSON and generates valid Dockerfiles from them.

## Installation

Add octopack to your Go project:

```bash
go get github.com/raghavyuva/octopack/packer
```

## Building

Build the tool using Go:

```bash
go build -o octopack ./cmd/main.go
```

## Usage

### Basic Usage

Read JSON from stdin and output Dockerfile to stdout:

```bash
cat input.json | ./octopack
```

### Read from File

Read JSON from a file and output to stdout:

```bash
./octopack -input testdata/simple-example.json
```

### Write to File

Read JSON from a file and write Dockerfile to a file:

```bash
./octopack -input testdata/simple-example.json -output Dockerfile
```

### Disable Validation

Skip validation (not recommended):

```bash
./octopack -input input.json -validate=false
```

## Command Line Options

- `-input`: Path to JSON input file (default: stdin)
- `-output`: Path to output Dockerfile file (default: stdout)
- `-validate`: Validate the JSON specification before generating (default: true)

## Example

Given a JSON file like `testdata/simple-example.json`:

```json
{
  "directives": {
    "syntax": "docker/dockerfile:1"
  },
  "stages": [
    {
      "from": {
        "image": "alpine",
        "tag": "latest"
      },
      "instructions": [
        {
          "type": "workdir",
          "path": "/app"
        },
        {
          "type": "run",
          "command": "apk add --no-cache python3"
        }
      ]
    }
  ]
}
```

Run:

```bash
./octopack -input testdata/simple-example.json -output Dockerfile
```

This will generate a valid Dockerfile based on your JSON specification.

## Using as a Library

Octopack can be used as a Go library in your applications. Import the `packer` package:

```go
import "github.com/raghavyuva/octopack/packer"
```

### Quick Start: Process JSON

The simplest way to use octopack as a library is with the `ProcessJSON` function:

```go
jsonData := []byte(`{
  "stages": [{
    "from": {"image": "alpine", "tag": "latest"},
    "instructions": [
      {"type": "run", "command": "echo 'Hello World'"}
    ]
  }]
}`)

dockerfile, err := packer.ProcessJSON(jsonData, true)
if err != nil {
    log.Fatal(err)
}
fmt.Println(dockerfile)
```

### Step by Step: Full Control

For more control, you can use the individual functions:

```go
// 1. Parse JSON into a DockerfileSpec
spec, err := packer.FromJSON(jsonData)
if err != nil {
    log.Fatal(err)
}

// 2. Validate the specification
result := packer.Validate(spec)
if result.HasErrors() {
    for _, err := range result.Errors {
        fmt.Printf("Error: %s\n", err)
    }
    return
}

// 3. Generate the Dockerfile
dockerfile, err := packer.Generate(spec)
if err != nil {
    log.Fatal(err)
}
```

### Programmatic Construction

You can also build DockerfileSpecs programmatically:

```go
spec := &packer.DockerfileSpec{
    Directives: &packer.ParserDirectives{
        Syntax: "docker/dockerfile:1",
    },
    Stages: []packer.Stage{
        {
            From: packer.FromInstruction{
                Image: "golang",
                Tag:   "1.21",
            },
            Instructions: []packer.Instruction{
                packer.WorkdirInstruction{Path: "/app"},
                packer.RunInstruction{
                    BaseInstruction: packer.BaseInstruction{
                        Command: []interface{}{"go", "build", "-o", "app", "."},
                    },
                },
                packer.EnvInstruction{
                    Variables: map[string]string{
                        "GO_ENV": "production",
                    },
                },
            },
        },
    },
}

dockerfile, err := packer.Generate(spec)
```

### Reading from io.Reader

You can also read JSON from any `io.Reader`:

```go
file, err := os.Open("spec.json")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

spec, err := packer.FromJSONReader(file)
if err != nil {
    log.Fatal(err)
}

// Or use ProcessJSONReader for a one-liner
dockerfile, err := packer.ProcessJSONReader(file, true)
```


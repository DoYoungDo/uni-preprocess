# uni-preprocess-go

A Go implementation of conditional compilation preprocessor for JavaScript and HTML files.

## Features

- Support for `#ifdef` and `#ifndef` directives
- JavaScript (`//`, `/* */`) and HTML (`<!-- -->`) comment styles
- Context-based variable evaluation
- Simple expression support (`==`, `!=`)

## Usage

```go
package main

import (
    "fmt"
    unipreprocess "uni-preprocess-go"
)

func main() {
    source := `// #ifdef DEBUG
console.log("Debug mode");
// #endif
console.log("Production code");`

    context := unipreprocess.ProcessContext{
        "DEBUG": true,
    }

    options := unipreprocess.PreprocessOptions{
        Type:    "js",
        Context: context,
    }

    result := unipreprocess.Preprocess(source, options)
    fmt.Println(result.Code)
}
```

## API

### Types

- `ProcessContext` - Map of variables for conditional evaluation
- `PreprocessOptions` - Configuration options
- `PreprocessResult` - Result containing processed code

### Functions

- `Preprocess(source string, options PreprocessOptions) PreprocessResult`

## Testing

```bash
go test
go test -v
```

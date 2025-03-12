# go-pantheon

**go-pantheon** is a game server framework that is ready to use. It provides a general server framework for microservices, allowing you to quickly build a high-performance and highly available game server cluster.

## vulcan-util

**vulcan-util** is a domain-agnostic utility library developed for the **go-pantheon** framework. This lightweight collection of modular components provides essential infrastructure tools including:

- Randomization and ID generation utilities
- Common data structure implementations
- Distributed systems primitives
- Time/date formatting extensions
- Hashing and encoding helpers
- Information security utilities

Designed for high performance in game server environments while maintaining framework independence.

## Usage

```go
import (
    "github.com/go-pantheon/vulcan-util/rand"
)

func main() {
    rand.RandAlphaNumString(10)
}
```

## Contributing

If you have any suggestions or feedback, please feel free to open an issue or submit a pull request.

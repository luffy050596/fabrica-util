# Fabrica Util

Fabrica Util is a comprehensive utility library for the go-pantheon ecosystem, providing common functionality for all go-pantheon components. This library encapsulates reusable code patterns, algorithms, and helper functions to ensure consistency and avoid duplication in the game server microservices infrastructure.

## go-pantheon Ecosystem

**go-pantheon** is an out-of-the-box game server framework providing high-performance, highly available game server cluster solutions based on microservices architecture based on [go-kratos](https://github.com/go-kratos/kratos). Fabrica Util serves as the foundational utility library that supports the core components:

- **Roma**: Game core logic services
- **Janus**: Gateway service for client connection handling and request forwarding
- **Lares**: Account service for user authentication and account management
- **Senate**: Backend management service providing operational interfaces

### Core Features

- 🔧 Common utility functions for time handling, synchronization, and randomization
- 🔐 Cryptographic security utilities including AES, RSA, and Curve25519
- 🧮 High-performance data structures (Bloom filter, Bitmap, Consistent Hash)
- 🔄 Concurrency utilities with thread-safe operations
- 📊 Data compression and manipulation utilities
- 🆔 Distributed ID generation tools
- 🌐 String manipulation including camelCase conversion

## Utility Packages

Fabrica Util provides a wide range of utility packages:

### Time Utilities (xtime/)
- Time format conversion
- Daily/weekly/monthly time calculations
- Multi-language time support

### Synchronization (xsync/)
- Thread-safe data structures
- Concurrency control primitives

### Randomization (xrand/)
- Secure random number generation
- Random string creation

### Security (security/)
- AES encryption/decryption
- RSA public/private key operations
- Curve25519 cryptography
- Secure channel implementation

### Data Structures
- Bloom filter (bloom/) for efficient set membership testing
- Bitmap (bitmap/) for memory-efficient bit operations
- Consistent Hash (consistenthash/) for distributed systems

### Other Utilities
- String manipulation (camelcase/)
- Data compression (compress/)
- ID generation (id/) for distributed systems, providing ID concatenation and obfuscation

## Technology Stack

| Technology/Component | Purpose | Version |
|---------|------|------|
| Go | Primary development language | 1.23+ |
| go-kratos | Microservice framework dependency | v2.8.4 |
| carbon | Time handling library | v2.6.2 |
| go-redis | Redis client for caching and rate limiting | v9.7.3 |
| atomic | Thread-safe atomic operations | v1.11.0 |
| crypto | Cryptographic operations | v0.37.0 |
| murmur3 | Hash algorithm | v1.1.0 |

## Requirements

- Go 1.23+

## Quick Start

### Installation

```bash
go get github.com/go-pantheon/fabrica-util
```

### Initialize Development Environment

```bash
make init
```

### Run Tests

```bash
make test
```

## Usage Examples

### Time Handling

```go
package main

import (
    "fmt"
    "time"

    "github.com/go-pantheon/fabrica-util/xtime"
)

func main() {
    // Initialize with language
    xtime.Init("en")

    // Format time
    fmt.Println(xtime.Format(time.Now()))

    // Get next daily reset time
    nextReset := xtime.NextDailyTime(time.Now(), 5 * time.Hour)
    fmt.Println("Next daily reset:", nextReset)
}
```

### AES Encryption

```go
package main

import (
    "fmt"

    "github.com/go-pantheon/fabrica-util/security/aes"
)

func main() {
    key := []byte("0123456789abcdef0123456789abcdef") // 32-byte key
    data := []byte("sensitive data")

    // Create cipher block
    block, err := aes.NewBlock(key)
    if err != nil {
        panic(err)
    }

    // Encrypt
    encrypted, err := aes.Encrypt(key, block, data)
    if err != nil {
        panic(err)
    }

    // Decrypt
    decrypted, err := aes.Decrypt(key, block, encrypted)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Original data: %s\n", data)
    fmt.Printf("Decrypted data: %s\n", decrypted)
}
```

### ID Obfuscation

```go
package main

import (
    "fmt"

    "github.com/go-pantheon/fabrica-util/id"
)

func main() {
    // Combine zone ID and entity ID
    zoneId := int64(1001)
    zone := uint8(5)
    combinedId := id.CombineZoneId(zoneId, zone)

    // Encrypt ID for frontend display
    encodedId, err := id.EncodeId(combinedId)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Encrypted ID: %s\n", encodedId)

    // Decrypt ID
    decodedId, err := id.DecodeId(encodedId)
    if err != nil {
        panic(err)
    }

    // Split ID to get original zone ID and entity ID
    originalZoneId, originalZone := id.SplitId(decodedId)
    fmt.Printf("Original zone ID: %d, Zone number: %d\n", originalZoneId, originalZone)
}
```

## Project Structure

```
.
├── xtime/              # Time utilities
├── xsync/              # Synchronization utilities
├── xrand/              # Random number generation
├── security/           # Cryptographic operations
│   ├── aes/            # AES encryption
│   ├── rsa/            # RSA encryption
│   ├── curve25519/     # Curve25519 cryptography
│   └── channel/        # Secure communication channels
├── consistenthash/     # Consistent hash implementation
├── data/               # Data handling utilities
├── id/                 # ID generation and obfuscation
├── bloom/              # Bloom filter implementation
├── compress/           # Data compression utilities
├── bitmap/             # Bitmap data structure
└── camelcase/          # String case conversion
```

## Integration with go-pantheon Components

Fabrica Util is designed to be imported by other go-pantheon components to provide common functionality:

```go
import (
    // Security utilities for token generation in Lares
    "github.com/go-pantheon/fabrica-util/security/aes"

    // Time utilities for game logic in Roma
    "github.com/go-pantheon/fabrica-util/xtime"

    // Synchronization utilities for connection handling in Janus
    "github.com/go-pantheon/fabrica-util/xsync"
)
```

## Development Guide

### License Compliance

The project enforces license compliance for all dependencies. We only allow the following licenses:
- MIT
- Apache-2.0
- BSD-2-Clause
- BSD-3-Clause
- ISC
- MPL-2.0

License checks are performed:
- Automatically in CI/CD pipelines
- Locally via pre-commit hooks
- Manually using `./hack/licenses-check`

For more information, see [License Check Documentation](hack/LICENSE-CHECK.md).

### Adding New Utilities

When adding new utility functions:

1. Create a new package or add to an existing one based on functionality
2. Implement the utility functions with proper error handling
3. Write comprehensive unit tests
4. Document usage with examples
5. Run tests: `make test`
6. Update documentation if needed

### Contribution Guidelines

1. Fork this repository
2. Create a feature branch
3. Submit changes with comprehensive tests
4. Ensure all tests pass
5. Submit a Pull Request

## License

This project is licensed under the terms specified in the LICENSE file.

# Zikrr

Zikrr is a powerful command-line tool for efficiently cloning multiple repositories from GitHub organizations. It provides an interactive terminal user interface (TUI) for selecting and managing repository cloning operations.

## Features

- 🔐 Secure GitHub authentication with support for both classic and fine-grained PATs
- 🎯 Interactive repository selection with filtering capabilities
- 📊 Real-time progress tracking for clone operations
- 🚀 Pagination support for handling large organizations
- ⚡ Rate limit aware GitHub API integration
- 🎨 Beautiful terminal UI powered by bubbletea

## Installation

### Prerequisites

- Go 1.21 or higher
- Git
- GitHub Personal Access Token (classic or fine-grained)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/sachin-duhan/zikrr.git
cd zikrr

# Build the binary
make build

# Or using go directly
go build -o zikrr ./cmd/zikrr
```

## Usage

### Basic Usage

```bash
# Using environment variable for token
export GITHUB_TOKEN=your_token_here
./zikrr

# Or passing token directly
./zikrr --token your_token_here
```

### Command Line Options

```bash
./zikrr [flags]

Flags:
  --token string      GitHub Personal Access Token
  --org string        GitHub Organization name (optional)
  --log-level string  Log level (debug, info, warn, error) (default "info")
```

### Interactive UI

1. **Organization Selection**: Enter the GitHub organization name you want to clone repositories from
2. **Repository Selection**: Browse and select repositories using:
   - ↑/↓: Navigate repositories
   - Space: Toggle repository selection
   - Enter: Confirm selection
   - /: Filter repositories
   - q: Quit
3. **Progress View**: Monitor cloning progress with real-time status updates

## Configuration

Zikrr can be configured using environment variables or command line flags:

```bash
# Environment Variables
GITHUB_TOKEN=your_token_here
ZIKRR_LOG_LEVEL=debug
ZIKRR_ORG=your-org-name
```

## Development

### Project Structure

```
zikrr/
├── cmd/zikrr/          # Main application entry point
├── internal/           # Internal packages
│   ├── auth/          # Authentication handling
│   ├── cli/           # CLI and TUI components
│   ├── config/        # Configuration management
│   ├── git/           # Git operations
│   └── github/        # GitHub API client
└── pkg/               # Public packages
```

### Building and Testing

```bash
# Run tests
make test

# Build binary
make build

# Clean build artifacts
make clean
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Add your license here]

# Zikrr - Project Structure

Below is the detailed project structure for Zikrr, a GitHub organization cloning tool written in Go. This structure follows Go best practices with a clear separation of concerns.

```
zikrr/
│
├── cmd/
│   └── zikrr/
│       └── main.go               # Application entry point
│
├── internal/                     # Internal packages
│   ├── auth/
│   │   ├── credentials.go        # GitHub authentication handling
│   │   └── token.go              # Token management
│   │
│   ├── cli/
│   │   ├── app.go                # CLI application setup
│   │   ├── commands.go           # Command definitions
│   │   ├── flags.go              # Flag definitions
│   │   └── interactive.go        # Interactive UI components
│   │
│   ├── config/
│   │   ├── config.go             # Configuration structure
│   │   └── loader.go             # Configuration loading/saving
│   │
│   ├── github/
│   │   ├── client.go             # GitHub API client
│   │   ├── organization.go       # Organization operations
│   │   └── repository.go         # Repository operations/models
│   │
│   └── git/
│       ├── clone.go              # Repository cloning
│       └── branch.go             # Branch operations
│
├── pkg/                          # Public packages
│   └── util/
│       ├── logger.go             # Logging utilities
│       └── progress.go           # Progress tracking
│
├── assets/                       # Project assets
│   └── banner.txt                # CLI banner/logo
│
├── .github/                      # GitHub specific files
│   ├── workflows/                # GitHub Actions workflows
│   │   └── build.yml             # CI/CD configuration
│   │
│   └── ISSUE_TEMPLATE/           # Issue templates
│       ├── bug_report.md
│       └── feature_request.md
│
├── scripts/                      # Build and utility scripts
│   ├── build.sh                  # Build script
│   └── install.sh                # Installation script
│
├── .gitignore                    # Git ignore file
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
├── LICENSE                       # License file
└── README.md                     # Project documentation
```

## Key Components Description

### CMD Package

The `cmd` package contains the entry point for the application. This is where the CLI application is initialized and run.

### Internal Packages

These packages are internal to the application and not intended to be imported by other projects.

#### Auth Package

Handles GitHub authentication, including token management and validation.

#### CLI Package

Contains all command-line interface components, including command definitions, flags, and interactive UI elements.

#### Config Package

Manages application configuration, including loading from files and environment variables.

#### GitHub Package

Provides integration with the GitHub API, including organization and repository operations.

#### Git Package

Handles Git operations, primarily cloning repositories and fetching branches.

### Public Packages (pkg)

These packages could potentially be used by other applications.

#### Util Package

Contains utility functions for logging and progress tracking.

### Additional Files

- **assets**: Contains static files used by the application
- **.github**: GitHub-specific files for CI/CD and issue templates
- **scripts**: Build and installation scripts
- **LICENSE**: License file (MIT recommended)
- **README.md**: Project documentation
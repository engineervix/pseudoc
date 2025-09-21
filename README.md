# pseudoc - Lorem Ipsum for Documents

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/engineervix/pseudoc)
[![CI/CD](https://github.com/engineervix/pseudoc/actions/workflows/main.yml/badge.svg)](https://github.com/engineervix/pseudoc/actions/workflows/main.yml)

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Features](#features)
- [Installation](#installation)
  - [Pre-compiled Binaries](#pre-compiled-binaries)
  - [From Source](#from-source)
- [Usage](#usage)
  - [Command-Line Interface (CLI)](#command-line-interface-cli)
    - [Commands and Options](#commands-and-options)
    - [Examples](#examples)
  - [HTTP API Server](#http-api-server)
    - [Starting the Server](#starting-the-server)
    - [Server Environments](#server-environments)
    - [API Endpoints](#api-endpoints)
    - [API Documentation](#api-documentation)
    - [API Examples](#api-examples)
    - [Server Configuration](#server-configuration)
- [Development](#development)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

`pseudoc` is a tool for generating placeholder documents to support testing and development workflows. Similar to how [Lorem Ipsum](https://en.wikipedia.org/wiki/Lorem_ipsum) provides placeholder text and [picsum.photos](https://picsum.photos) provides placeholder images, pseudoc generates various types of fake documents to aid in development and testing scenarios.

## Features

- **Multiple Document Formats**: Generate PDF, DOCX, and XLSX files.
- **Dual Interface**: Use it as a command-line tool or as an HTTP server.
- **Customizable Output**: Control the number of documents, pages, sheets, and filenames.
- **Reproducible Generation**: Use a seed to generate the same documents every time.
- **Cross-Platform**: A single, self-contained binary for all major platforms (Windows, macOS, Linux).
- **Lightweight & Fast**: Minimal dependencies and optimized for quick document generation.

## Installation

### Pre-compiled Binaries

Pre-compiled binaries for all major platforms are available on the [GitHub Releases](https://github.com/engineervix/pseudoc/releases) page. Download the appropriate binary for your system, add it somewhere appropriate, e.g. a directory in your PATH, and you're ready to go.

### From Source

If you have Go installed, you can build `pseudoc` from source:

```sh
git clone https://github.com/engineervix/pseudoc.git
cd pseudoc
go build -o bin/pseudoc ./cmd/pseudoc
```

## Usage

`pseudoc` can be used as a command-line tool or as an HTTP server.

### Command-Line Interface (CLI)

The CLI is perfect for scripting and local document generation.

#### Commands and Options

```
USAGE:
  pseudoc [command] [options]

COMMANDS:
  pdf                          Generate PDF document
  docx, word                   Generate Word document
  xlsx, spreadsheet, sheet     Generate Excel spreadsheet
  random                       Generate random document type
  serve, server                Start HTTP API server
  version, -v, --version       Show version information
  help, -h, --help             Show this help message

OPTIONS:
  --count N                    Number of documents to generate (default: 1)
  --output-dir DIR             Directory to store output files (default: current directory)
  --filename NAME              Base filename (default: auto-generated with timestamp)
  --seed VALUE                 Set random seed for reproducible output
  --quiet                      Suppress output for scripting
  --dry-run                    Preview what would be generated without creating files

FORMAT-SPECIFIC OPTIONS:
  --pages N                    For PDF/DOCX: Number of pages (default: 1)
  --sheets N                   For XLSX: Number of sheets (default: 1)
```

#### Examples

- **Generate a single PDF:**

    ```sh
    pseudoc pdf
    ```

- **Generate 5 Word documents with 3 pages each:**

    ```sh
    pseudoc docx --count 5 --pages 3
    ```

- **Generate an Excel file with a custom filename and 4 sheets:**

    ```sh
    pseudoc xlsx --filename my-data --sheets 4
    ```

- **Generate 10 random documents in a specific directory:**

    ```sh
    pseudoc random --count 10 --output-dir ./test-files
    ```

- **Generate reproducible documents with a seed:**

    ```sh
    pseudoc random --seed 42 --count 5
    ```

### HTTP API Server

Start the server to generate documents via HTTP requests. This is useful for various kinds of integrations.

#### Starting the Server

```sh
pseudoc serve
```

By default, the server starts on `localhost:8080` in **development** mode. You can customize the host, port, and environment:

```sh
pseudoc serve --host 0.0.0.0 --port 3000 --env production
```

#### Server Environments

- **development (default)**: Optimized for local testing. Security features like HSTS are disabled to allow for easy testing over plain HTTP.
- **production**: Intended for deployment. Stricter security headers are enabled. The server should be run behind a TLS proxy (like Nginx or a cloud load balancer) in this mode.

#### API Endpoints

- `GET /health`: Health check endpoint.
- `GET /api/v1/info`: Returns server information and capabilities.
- `GET /api/v1/generate/{type}`: Generate a document.
- `POST /api/v1/generate`: Generate a document with a JSON config.
- `GET /metrics`: Server metrics endpoint (when enabled).
- `GET /docs`: Interactive API documentation (Swagger UI).

#### API Documentation

The server provides interactive API documentation via Swagger UI at `/docs`. This includes:

- **Interactive Testing**: Try out API endpoints directly from the browser
- **Request/Response Examples**: See example requests and responses for all endpoints
- **Schema Documentation**: Detailed documentation of all request and response formats
- **Parameter Validation**: Real-time validation of API parameters

Visit `http://localhost:8080/docs` when the server is running to explore the full API documentation.

#### API Examples

- **Generate a 3-page PDF using a GET request:**

    ```sh
    curl "http://localhost:8080/api/v1/generate/pdf?pages=3" -o document.pdf
    ```

- **Generate a random document with a specific seed:**

    ```sh
    curl "http://localhost:8080/api/v1/generate/random?seed=42" -o random-doc
    ```

- **Generate a 5-page DOCX using a POST request:**

    ```sh
    curl -X POST "http://localhost:8080/api/v1/generate" \
      -H "Content-Type: application/json" \
      -d '{"type":"docx","pages":5,"seed":123}' \
      -o document.docx
    ```

When using the `GET /api/v1/generate/random` endpoint, the server will generate a document of a randomly selected type directly. Both GET and POST requests include the `X-Pseudoc-Generated-Type` response header to indicate which document type was generated.

```sh
# Generate a random document - the type will be determined by the server
curl "http://localhost:8080/api/v1/generate/random?seed=42" -o random-doc
```

Use the `-L -J -O` flags to automatically save the file with the server-provided filename and extension:

```sh
# The server provides the correct filename via Content-Disposition header
curl -L -J -O "http://localhost:8080/api/v1/generate/random?seed=42"
```

You can check the generated document type from the response header:

```sh
curl -I "http://localhost:8080/api/v1/generate/random?seed=42"
# Look for: X-Pseudoc-Generated-Type: pdf
```

#### Server Configuration

The server supports various configuration options for production deployment:

- **CORS**: Disabled by default for security. Enable by setting allowed origins.
- **Rate Limiting**: 60 requests per minute by default. Set to 0 to disable.
- **Metrics**: Disabled by default. Enable with `--metrics` flag. Basic authentication can be configured programmatically.
- **Request Timeout**: 30 seconds by default.
- **Max File Size**: 100MB limit for generated documents.

Example with custom configuration:

```sh
# Enable metrics endpoint
pseudoc serve --metrics

# Custom rate limiting and CORS
pseudoc serve --rate-limit 120 --cors-allowed-origins "https://example.com,https://app.example.com"

# Production setup with custom timeout and logging
pseudoc serve --env production --timeout 60s --log-level warn
```

## Development

This project uses `just` as a command runner. See the [`justfile`](./justfile) for available commands.

- **Build the project:**
    ```sh
    just build
    ```
- **Run tests:**
    ```sh
    just test
    ```
- **Run the tool locally:**
    ```sh
    just run -- pdf --count 2
    ```
- **Generate API documentation:**
    ```sh
    just docs
    ```

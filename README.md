# gomark

[![Go Report Card](https://goreportcard.com/badge/github.com/mlctrez/gomark)](https://goreportcard.com/report/github.com/mlctrez/gomark)

`gomark` is a command-line tool written in Go that converts Markdown files into self-contained HTML documents. It features a modern dark theme, syntax highlighting for code blocks, and support for GitHub Flavored Markdown (GFM).

## Features

- **Self-contained HTML:** All styles and syntax highlighting CSS are embedded directly in the output HTML file.
- **Dark Theme:** Modern dark interface by default.
- **Syntax Highlighting:** High-quality code block highlighting using Chroma (Monokai style).
- **GFM Support:** Includes support for tables, task lists, and other GitHub Flavored Markdown extensions.
- **Wide Layout:** Content width optimized for readability (1200px max-width).

## Installation

To install `gomark`, ensure you have Go installed on your system, then run:

```bash
go install github.com/mlctrez/gomark@latest
```

The `gomark` binary will be placed in your `$GOPATH/bin` (usually `~/go/bin`).

## Usage

### Quick Preview

Pass a Markdown file as the only argument to render it to a temporary HTML file and open it in your default browser:

```bash
gomark README.md
```

The HTML file is written to the OS temp directory (e.g. `/tmp/README-123456.html`) and opened automatically.

### Explicit Output

Use the `-i` and `-o` flags to specify both input and output paths:

```bash
gomark -i <input.md> -o <output.html>
```

#### Flags

- `-i`: Path to the input Markdown file.
- `-o`: Path to the output HTML file.

### Examples

```bash
# Preview in browser
gomark README.md

# Convert to a specific file
gomark -i README.md -o documentation.html
```

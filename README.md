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

Run `gomark` with the input and output file paths:

```bash
gomark -i <input.md> -o <output.html>
```

### Flags

- `-i`: Path to the input Markdown file.
- `-o`: Path to the output HTML file.

## Example

```bash
gomark -i README.md -o documentation.html
```

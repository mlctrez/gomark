package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	rendererhtml "github.com/yuin/goldmark/renderer/html"
)

//go:embed templates/layout.html
var layoutHTML string

func main() {
	inputPath := flag.String("i", "", "Input markdown file")
	outputPath := flag.String("o", "", "Output html file")
	flag.Parse()

	var inFile, outFile string

	switch {
	case *inputPath != "" && *outputPath != "":
		// Explicit flags: gomark -i input.md -o output.html
		inFile = *inputPath
		outFile = *outputPath
	case *inputPath == "" && *outputPath == "" && flag.NArg() == 1:
		// Positional arg only: gomark README.md → temp file + open browser
		inFile = flag.Arg(0)
		base := strings.TrimSuffix(filepath.Base(inFile), filepath.Ext(inFile))
		tmpFile, err := os.CreateTemp("", base+"-*.html")
		if err != nil {
			log.Fatalf("failed to create temp file: %v", err)
		}
		outFile = tmpFile.Name()
		tmpFile.Close()
	default:
		fmt.Println("Usage: gomark <input.md>")
		fmt.Println("       gomark -i <input.md> -o <output.html>")
		os.Exit(1)
	}

	openBrowser := *inputPath == "" && *outputPath == ""

	mdContent, err := os.ReadFile(inFile)
	if err != nil {
		log.Fatalf("failed to read input file: %v", err)
	}

	htmlBytes, err := renderMarkdown(mdContent)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(outFile, htmlBytes, 0644); err != nil {
		log.Fatalf("failed to write output file: %v", err)
	}

	if openBrowser {
		fmt.Printf("Opening %s\n", outFile)
		if err := browserOpen(outFile); err != nil {
			log.Fatalf("failed to open browser: %v", err)
		}
	} else {
		fmt.Printf("Successfully converted %s to %s\n", inFile, outFile)
	}
}

func renderMarkdown(mdContent []byte) ([]byte, error) {
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					html.WithLineNumbers(false),
					html.WithClasses(true),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			rendererhtml.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := gm.Convert(mdContent, &buf); err != nil {
		return nil, fmt.Errorf("failed to convert markdown: %w", err)
	}

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}
	formatter := html.New(html.WithAllClasses(true))
	var cssBuf bytes.Buffer
	if err := formatter.WriteCSS(&cssBuf, style); err != nil {
		return nil, fmt.Errorf("failed to write CSS: %w", err)
	}

	tmpl, err := template.New("layout").Parse(layoutHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	data := struct {
		HighlightingCSS template.CSS
		Content         template.HTML
	}{
		HighlightingCSS: template.CSS(cssBuf.String()),
		Content:         template.HTML(buf.String()),
	}

	var finalBuf bytes.Buffer
	if err := tmpl.Execute(&finalBuf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return finalBuf.Bytes(), nil
}

func browserOpen(url string) error {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "start"
	default:
		cmd = "xdg-open"
	}
	return exec.Command(cmd, url).Start()
}

package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"

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

	if *inputPath == "" || *outputPath == "" {
		fmt.Println("Usage: gomark -i <input.md> -o <output.html>")
		os.Exit(1)
	}

	mdContent, err := ioutil.ReadFile(*inputPath)
	if err != nil {
		log.Fatalf("failed to read input file: %v", err)
	}

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
		log.Fatalf("failed to convert markdown: %v", err)
	}

	// Get CSS for highlighting
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}
	formatter := html.New(html.WithAllClasses(true))
	var cssBuf bytes.Buffer
	err = formatter.WriteCSS(&cssBuf, style)
	if err != nil {
		log.Fatalf("failed to write CSS: %v", err)
	}

	tmpl, err := template.New("layout").Parse(layoutHTML)
	if err != nil {
		log.Fatalf("failed to parse template: %v", err)
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
		log.Fatalf("failed to execute template: %v", err)
	}

	err = ioutil.WriteFile(*outputPath, finalBuf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("failed to write output file: %v", err)
	}

	fmt.Printf("Successfully converted %s to %s\n", *inputPath, *outputPath)
}

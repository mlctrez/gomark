package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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

//go:embed templates/index.html
var indexHTML string

func main() {
	inputPath := flag.String("i", "", "Input markdown file")
	outputPath := flag.String("o", "", "Output html file")
	flag.Parse()

	switch {
	case *inputPath != "" && *outputPath != "":
		convertFile(*inputPath, *outputPath)
	case *inputPath == "" && *outputPath == "" && flag.NArg() == 1:
		arg := flag.Arg(0)
		info, err := os.Stat(arg)
		if err != nil {
			log.Fatalf("cannot access %s: %v", arg, err)
		}
		if info.IsDir() {
			serveDirectory(arg)
		} else {
			previewFile(arg)
		}
	default:
		fmt.Println("Usage: gomark <input.md>")
		fmt.Println("       gomark <directory>")
		fmt.Println("       gomark -i <input.md> -o <output.html>")
		os.Exit(1)
	}
}

func convertFile(inFile, outFile string) {
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
	fmt.Printf("Successfully converted %s to %s\n", inFile, outFile)
}

func previewFile(inFile string) {
	base := strings.TrimSuffix(filepath.Base(inFile), filepath.Ext(inFile))
	tmpFile, err := os.CreateTemp("", base+"-*.html")
	if err != nil {
		log.Fatalf("failed to create temp file: %v", err)
	}
	outFile := tmpFile.Name()
	tmpFile.Close()

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
	fmt.Printf("Opening %s\n", outFile)
	if err := browserOpen(outFile); err != nil {
		log.Fatalf("failed to open browser: %v", err)
	}
}

func serveDirectory(dir string) {
	mdFiles, err := findMarkdownFiles(dir)
	if err != nil {
		log.Fatalf("failed to scan directory: %v", err)
	}
	if len(mdFiles) == 0 {
		log.Fatalf("no markdown files found in %s", dir)
	}

	indexTmpl, err := template.New("index").Parse(indexHTML)
	if err != nil {
		log.Fatalf("failed to parse index template: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Rescan on each request so new files are picked up
		files, err := findMarkdownFiles(dir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := struct{ Files []string }{Files: files}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		indexTmpl.Execute(w, data)
	})

	http.HandleFunc("/view/", func(w http.ResponseWriter, r *http.Request) {
		relPath := strings.TrimPrefix(r.URL.Path, "/view/")
		if relPath == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		// Prevent path traversal
		cleaned := filepath.Clean(relPath)
		if strings.HasPrefix(cleaned, "..") {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		fullPath := filepath.Join(dir, cleaned)
		mdContent, err := os.ReadFile(fullPath)
		if err != nil {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		htmlBytes, err := renderMarkdown(mdContent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(htmlBytes)
	})

	addr := ":9090"
	url := "http://localhost" + addr
	fmt.Printf("Serving markdown files from %s at %s\n", dir, url)

	go func() {
		time.Sleep(200 * time.Millisecond)
		if err := browserOpen(url); err != nil {
			log.Printf("failed to open browser: %v", err)
		}
	}()

	log.Fatal(http.ListenAndServe(addr, nil))
}

func findMarkdownFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".md" || ext == ".markdown" {
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})
	return files, err
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

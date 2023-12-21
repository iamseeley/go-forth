package static

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/go-yaml/yaml"
	"github.com/russross/blackfriday/v2"
	"go-forth2.0/internal/config"
)

// Define a data structure for your site pages
type Page struct {
	Title       string
	Description string
	Body        []byte
}

// Load site pages written in Markdown from a directory
func loadPageFromDirectory(directory, title string) (*Page, error) {
	filename := directory + title + ".md"
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	frontMatter, body, err := parseFrontMatter(content)
	if err != nil {
		return nil, err
	}

	var page Page
	if title, ok := frontMatter["title"].(string); ok {
		page.Title = title
	}
	if description, ok := frontMatter["description"].(string); ok {
		page.Description = description
	}
	page.Body = body

	return &page, nil
}

func parseFrontMatter(content []byte) (map[string]interface{}, []byte, error) {
	frontMatter := make(map[string]interface{})
	var contentStart int

	delimiter := []byte("---")
	start := bytes.Index(content, delimiter)
	if start == -1 {
		return nil, nil, errors.New("Front matter delimiter not found")
	}

	end := bytes.Index(content[start+len(delimiter):], delimiter)
	if end == -1 {
		return nil, nil, errors.New("Second front matter delimiter not found")
	}

	if err := yaml.Unmarshal(content[start+len(delimiter):start+len(delimiter)+end], &frontMatter); err != nil {
		return nil, nil, err
	}

	contentStart = start + len(delimiter) + end + len(delimiter)
	actualContent := content[contentStart:]

	return frontMatter, actualContent, nil
}

// ... existing code from site.go ...

// BuildSite generates static HTML files from Markdown content
func BuildSite() {
	cfg, err := config.LoadConfig("./config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	contentDir := cfg.ContentPath // Directory containing Markdown files
	outputDir := "output/"        // Directory to save generated HTML files

	// Ensure output directory exists
	os.MkdirAll(outputDir, os.ModePerm)

	// Iterate through all Markdown files in the content directory
	err = filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".md" {
			// Generate HTML for each Markdown file
			return generateHTML(path, outputDir)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error building site: %v", err)
	}

	log.Println("Site built successfully")
}

func markDowner(args ...interface{}) template.HTML {
	s := blackfriday.Run([]byte(fmt.Sprintf("%s", args...)))
	return template.HTML(s)
}

var templates = template.Must(template.New("").Funcs(template.FuncMap{"markDown": markDowner}).ParseGlob("templates/*.html"))

// generateHTML converts a Markdown file to HTML and saves it
func generateHTML(mdPath, outputDir string) error {
	fileName := filepath.Base(mdPath)
	baseName := fileName[:len(fileName)-len(filepath.Ext(fileName))]

	page, err := loadPageFromDirectory(filepath.Dir(mdPath)+"/", baseName)
	if err != nil {
		return err
	}

	var rendered bytes.Buffer
	err = templates.ExecuteTemplate(&rendered, "site", page) // Adjust "site" to your template name
	if err != nil {
		return err
	}

	// Save the rendered HTML
	return os.WriteFile(filepath.Join(outputDir, baseName+".html"), rendered.Bytes(), 0644)
}

// ... remaining server code ...

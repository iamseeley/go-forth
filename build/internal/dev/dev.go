package dev

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"

	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/go-yaml/yaml"
	"github.com/russross/blackfriday/v2"
	"go-forth2.0/build/internal/config"
)

// Now you can use cfg to access your configuration settings

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

func pageHandler(w http.ResponseWriter, r *http.Request, title string) {

	cfg, err := config.LoadConfig("./config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	setCacheHeaders(w, 600)
	p, err := loadPageFromDirectory(cfg.ContentPath, title)
	if err != nil {
		// Instead of redirecting, send an error message or render an error template
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	renderTemplate(w, "site", p)
	// fmt.Fprintf(w, "Hello, you've requested: %s", title)
}

func markDowner(args ...interface{}) template.HTML {
	s := blackfriday.Run([]byte(fmt.Sprintf("%s", args...)))
	return template.HTML(s)
}

// var templates = template.Must(template.New("").Funcs(template.FuncMap{"markDown": markDowner}).ParseGlob("templates/*.html"))

var templates *template.Template

func loadTemplates() error {
	var err error
	templates, err = template.New("").Funcs(template.FuncMap{"markDown": markDowner}).ParseFiles("templates/site.html")
	if err != nil {
		return fmt.Errorf("error loading site template: %w", err)
	}
	return nil
}

func init() {
	err := loadTemplates()
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, content interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, content)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}

		fn(w, r, m[1])
	}
}

func setCacheHeaders(w http.ResponseWriter, maxAge int) {
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", maxAge))
}

func StartServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			pageHandler(w, r, "index") // Serve index.md for the root path
		} else {
			makeHandler(pageHandler)(w, r) // Continue with regular routing for other paths
		}
	})
	// http.Handle("/", http.RedirectHandler("/index", http.StatusSeeOther))
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))
	// http.HandleFunc("/", makeHandler(pageHandler))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

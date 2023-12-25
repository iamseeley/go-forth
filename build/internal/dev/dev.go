package dev

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
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
	Theme       string
}

func watchContentDirectory(contentDir, templateDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(contentDir)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					createTemplateForDir(filepath.Base(event.Name), templateDir)
				}
			}
		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}

func createTemplateForDir(dirName, templateDir string) {
	templatePath := filepath.Join(templateDir, dirName+".html")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// Basic HTML template content
		templateContent := fmt.Sprintf(`
{{ define "%s" }}
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Go-Forth - {{.Page.Title}}</title>
	<link type="text/css" rel="stylesheet" href="/public/css/{{.Page.Theme}}.css">
</head>
<body>
	<header><nav><div></div><a href="/">home</a><a href="/">post</a></nav></header>
	
	<main>
		<h2>{{.Page.Title}}</h2>
		{{ .Page.Body | markDown }}
		{{ range .Data.projects }}
			<div>
				<h3>{{ .name }}</h3>
				<p>{{ .description }}</p>
			</div>
		{{ end }}

	</main>
	<footer>made by Thomas</footer>
</body>
</html>
{{ end }}
`, dirName)

		err = os.WriteFile(templatePath, []byte(templateContent), 0644)
		if err != nil {
			log.Printf("Failed to create template for %s: %v", dirName, err)
		} else {
			log.Printf("Created template: %s", templatePath)
		}
	}
}

// Load site pages written in Markdown from a directory
func loadPageFromDirectory(directory, title string) (*Page, error) {
	filename := directory + title
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

	cfg, err := config.LoadConfig("./config.json") // Load configuration
	if err != nil {
		return nil, err
	}

	page.Theme = cfg.ThemeName
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

func loadData(directory string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".json" {
			fileData, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var jsonData interface{}
			if err := json.Unmarshal(fileData, &jsonData); err != nil {
				return err
			}

			key := filepath.Base(path)
			key = strings.TrimSuffix(key, filepath.Ext(key)) // Use filename as the key
			data[key] = jsonData
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

func pageHandler(w http.ResponseWriter, r *http.Request, filePath string) {
	cfg, err := config.LoadConfig("./config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	setCacheHeaders(w, 600)

	log.Printf("Constructed file path: %s", filePath)

	data, err := loadData("data")
	if err != nil {
		log.Printf("Failed to load data: %v", err)
	}

	p, err := loadPageFromDirectory(cfg.ContentPath, filePath)
	if err != nil {
		log.Printf("Error loading page: %v", err) // Log the error for debugging
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	collection := filepath.Base(filepath.Dir(filePath))

	templateData := struct {
		Page *Page
		Data map[string]interface{}
	}{
		Page: p,
		Data: data,
	}

	tmplName := collection
	tmpl := templates.Lookup(tmplName + ".html")
	if tmpl == nil {
		log.Printf("Template %s.html not found, using default site.html", tmplName)
		tmpl = templates.Lookup("site.html")
	}

	renderTemplate(w, tmplName, templateData)
}

func markDowner(args ...interface{}) template.HTML {
	s := blackfriday.Run([]byte(fmt.Sprintf("%s", args...)))
	return template.HTML(s)
}

// var templates = template.Must(template.New("").Funcs(template.FuncMap{"markDown": markDowner}).ParseGlob("templates/*.html"))

var templates *template.Template

func loadTemplates() error {
	var err error
	templates, err = template.New("").Funcs(template.FuncMap{"markDown": markDowner}).ParseGlob("templates/*.html")
	if err != nil {
		return fmt.Errorf("error loading templates: %w", err)
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

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func copyDir(src string, dst string) error {
	// Get properties of source dir
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create the destination directory
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(src)
	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		srcFile := filepath.Join(src, obj.Name())
		dstFile := filepath.Join(dst, obj.Name())

		if obj.IsDir() {
			// Create sub-directories - recursively
			err = copyDir(srcFile, dstFile)
			if err != nil {
				return err
			}
		} else {
			// Perform the file copy
			err = copyFile(srcFile, dstFile)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func StartServer() {
	// Load configuration
	cfg, err := config.LoadConfig("./config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Copy the theme CSS to the assets/css directory
	themeCSSPath := filepath.Join("themes", cfg.ThemeName+".css")
	assetsCSSPath := filepath.Join("assets/css", cfg.ThemeName+".css")
	os.MkdirAll(filepath.Dir(assetsCSSPath), os.ModePerm) // Create the assets/css directory
	err = copyFile(themeCSSPath, assetsCSSPath)
	if err != nil {
		log.Fatalf("Failed to copy theme CSS to assets directory: %v", err)
	}

	err = loadTemplates()
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	outputDir := cfg.OutputPath

	os.MkdirAll(outputDir, os.ModePerm)

	// Copy the assets directory to public in the output directory
	assetsSrc := "assets"
	assetsDst := filepath.Join(outputDir, "public")
	err = copyDir(assetsSrc, assetsDst)
	if err != nil {
		log.Fatalf("Failed to copy assets directory: %v", err)
	}

	go watchContentDirectory("content", "templates")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			// Serve 'index.md' from the 'page' directory for the root path
			pageHandler(w, r, "page/index.md")
		case !strings.Contains(r.URL.Path[1:], "/"):
			// Handle 'page' collection routes without the 'page' prefix in the URL
			// For example, "/about" will serve "page/about.md"
			strippedPath := strings.TrimPrefix(r.URL.Path, "/")
			pageHandler(w, r, fmt.Sprintf("page/%s.md", strippedPath))
		default:
			// Handle other collection routes
			// For example, "/post/post1" will serve "post/post1.md"
			pageHandler(w, r, r.URL.Path[1:]+".md") // [1:] to remove the leading '/'
		}
	})

	// http.Handle("/", http.RedirectHandler("/index", http.StatusSeeOther))
	fs := http.FileServer(http.Dir("src/public"))
	http.Handle("/public/", http.StripPrefix("/public/", fs))
	// http.HandleFunc("/", makeHandler(pageHandler))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

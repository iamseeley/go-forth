package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	// "path/filepath"
	"regexp"
	"github.com/go-yaml/yaml"
	"github.com/russross/blackfriday/v2"
)

// Define a data structure for your site pages!

type Page struct {
	Title       string
	Description string
	Body        []byte
}

// Add other page types! This is the journal struct I use in my personal site.

// type Journal struct {
// 	Title       string
// 	Body        []byte
// 	Image       string
// 	Description string
// 	Date        string
// 	NextUrl     string
// 	PrevUrl     string
// }



// Load site pages written in Markdown from a directory. 

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

	// Extract and set front matter data into the Page struct
	if title, ok := frontMatter["title"].(string); ok {
		page.Title = title
	}

	if description, ok := frontMatter["description"].(string); ok {
		page.Description = description
	}

	page.Body = body

	return &page, nil
}

// var AllJournals []Journal
// var JournalMap = make(map[string]int)
//
// func loadAllJournals(directory string) {
// 	AllJournals = []Journal{}
// 	files, err := os.ReadDir(directory)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	for i, file := range files {
// 		if filepath.Ext(file.Name()) == ".md" {
// 			title := file.Name()[0 : len(file.Name())-3]
// 			JournalMap[title] = i
// 			journal, err := loadJournalFromDirectory(directory, title)
// 			if err != nil {
// 				log.Printf("Could not load journal: %s", err)
// 				continue
// 			}
// 			AllJournals = append(AllJournals, *journal)
// 		}
// 	}
// }

// func loadJournalFromDirectory(directory, title string) (*Journal, error) {
// 	filename := directory + title + ".md"
// 	content, err := os.ReadFile(filename)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	frontMatter, body, err := parseFrontMatter(content)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var journal Journal
//
// 	// Extract and set front matter data into the Journal struct
// 	if title, ok := frontMatter["title"].(string); ok {
// 		journal.Title = title
// 	}
//
// 	if image, ok := frontMatter["image"].(string); ok {
// 		journal.Image = image
// 	}
//
// 	if description, ok := frontMatter["description"].(string); ok {
// 		journal.Description = description
// 	}
//
// 	if date, ok := frontMatter["date"].(string); ok {
// 		journal.Date = date
// 	}
//
// 	journal.Body = body
//
// 	return &journal, nil
// }

func parseFrontMatter(content []byte) (map[string]interface{}, []byte, error) {
	frontMatter := make(map[string]interface{})
	var contentStart int

	// Find the position of the first `---` delimiter
	delimiter := []byte("---")
	start := bytes.Index(content, delimiter)
	if start == -1 {
		return nil, nil, errors.New("Front matter delimiter not found")
	}

	// Find the position of the second `---` delimiter
	end := bytes.Index(content[start+len(delimiter):], delimiter)
	if end == -1 {
		return nil, nil, errors.New("Second front matter delimiter not found")
	}

	// Parse the front matter
	if err := yaml.Unmarshal(content[start+len(delimiter):start+len(delimiter)+end], &frontMatter); err != nil {
		return nil, nil, err
	}

	// Find the start of the actual content
	contentStart = start + len(delimiter) + end + len(delimiter)

	// Extract the front matter and content separately
	actualContent := content[contentStart:]

	return frontMatter, actualContent, nil
}

// Attempts to load a page by its title from the "pages/" directory.
// If successful, it renders the page using a template. If it encounters an error, it redirects the client to a different location. 

func pageHandler(w http.ResponseWriter, r *http.Request, title string) {
	setCacheHeaders(w, 600)
	p, err := loadPageFromDirectory("pages/", title)
	if err != nil {
		http.Redirect(w, r, "/site/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "site", p)
}

// func journalHandler(w http.ResponseWriter, r *http.Request, title string) {
// 	setCacheHeaders(w, 600)
// 	journal, err := loadJournalFromDirectory("journal/", title)
// 	if err != nil {
// 		http.Redirect(w, r, "/journal/"+title, http.StatusFound)
// 		return
// 	}
//
// 	currentIndex, exists := JournalMap[title]
// 	if !exists {
// 		http.Error(w, "Journal not found", http.StatusNotFound)
// 		return
// 	}
//
// 	if currentIndex > 0 {
// 		prevTitle := AllJournals[currentIndex-1].Title
// 		journal.PrevUrl = "/journal/" + prevTitle
// 	}
//
// 	if currentIndex < len(AllJournals)-1 {
// 		nextTitle := AllJournals[currentIndex+1].Title
// 		journal.NextUrl = "/journal/" + nextTitle
// 	}
//
// 	renderTemplate(w, "journal", journal)
// }

// This is the function that goes in the html/template and processes the Markdown.

func markDowner(args ...interface{}) template.HTML {
	s := blackfriday.Run([]byte(fmt.Sprintf("%s", args...)))
	return template.HTML(s)
}

// Any template you add to the templates directory will be parsed here.

var templates = template.Must(template.New("").Funcs(template.FuncMap{"markDown": markDowner}).ParseGlob("templates/*.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, content interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Add valid page paths here.
var validPath = regexp.MustCompile("^/(site)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}

		fn(w, r, m[2])
	}
}

func setCacheHeaders(w http.ResponseWriter, maxAge int) {
	// Set cache control headers to enable caching for the specified duration
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", maxAge))
}

func main() {

	// Redirect index to site/home
	http.Handle("/", http.RedirectHandler("/site/home", http.StatusSeeOther))

	// Serve the assets directory
	fs := http.FileServer(http.Dir("assets"))

	http.Handle("/assets/", http.StripPrefix("/assets/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set cache headers to cache assets for 1 hour (you can adjust this duration)
		setCacheHeaders(w, 600)
		fs.ServeHTTP(w, r)
	})))


	// Loads all pages to implement pagination.
	/* loadAllJournals("journal/") */
	
	// Routes to appropriate handler
	http.HandleFunc("/site/", makeHandler(pageHandler))

	/* http.HandleFunc("/journal/", makeHandler(journalHandler)) */

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

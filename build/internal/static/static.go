package static

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"go-forth2.0/build/internal/config"
	"go-forth2.0/build/internal/models"
	"go-forth2.0/build/internal/utils"
)

func init() {
	err := utils.LoadTemplates()
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}
}

// BuildSite generates static HTML files from Markdown content
func BuildSite() {
	cfg, err := config.LoadConfig("./config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	data, err := utils.LoadData(cfg.DataPath)
	if err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}

	// Copy the theme CSS to the assets/css directory
	themeCSSPath := filepath.Join("themes", cfg.ThemeName+".css")
	assetsCSSPath := filepath.Join("assets/css", cfg.ThemeName+".css")
	os.MkdirAll(filepath.Dir(assetsCSSPath), os.ModePerm) // Create the assets/css directory if it doesn't exist
	err = utils.CopyFile(themeCSSPath, assetsCSSPath)
	if err != nil {
		log.Fatalf("Failed to copy theme CSS to assets directory: %v", err)
	}

	contentDir := cfg.ContentPath
	outputDir := cfg.OutputPath

	os.MkdirAll(outputDir, os.ModePerm)

	// Copy the assets directory to public in the output directory
	assetsSrc := "assets"
	assetsDst := filepath.Join(outputDir, "public")
	err = utils.CopyDir(assetsSrc, assetsDst)
	if err != nil {
		log.Fatalf("Failed to copy assets directory: %v", err)
	}

	// Iterate through all Markdown files in the content directory
	err = filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".md" {
			return generateHTML(path, outputDir, data, cfg) // Pass the loaded data and cfg
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error building site: %v", err)
	}

	log.Println("Site built successfully")
}

func generateHTML(mdPath, outputDir string, data map[string]interface{}, cfg *config.Config) error {
	// Extract the relative path of the Markdown file from the content directory
	relativePath, err := filepath.Rel(cfg.ContentPath, mdPath)
	if err != nil {
		log.Printf("Error getting relative path: %v", err)
		return err
	}

	// Change the extension from .md to .html
	htmlPath := strings.TrimSuffix(relativePath, filepath.Ext(relativePath)) + ".html"

	// If the file is in the 'page' directory, place it at the root of the output directory
	if strings.HasPrefix(relativePath, "page"+string(filepath.Separator)) {
		htmlPath = strings.TrimPrefix(htmlPath, "page"+string(filepath.Separator))
	}

	// Create the full path for the output HTML file
	outputPath := filepath.Join(outputDir, htmlPath)

	// Create the necessary directories in the output path
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		log.Printf("Error creating directories: %v", err)
		return err
	}

	page, err := utils.LoadPageFromDirectory(filepath.Dir(mdPath)+"/", filepath.Base(mdPath))
	if err != nil {
		log.Printf("Error loading page: %v", err)

		return err // Log the error for debugging
	}

	// Generate the OG Image URL
	ogImageFileName := strings.TrimSuffix(filepath.Base(outputPath), filepath.Ext(outputPath)) + "-og-image.png"
	ogImageUrl := "/public/og-image/" + ogImageFileName
	page.OGImageURL = ogImageUrl

	// Determine template based on the collection (parent directory name)
	collection := filepath.Base(filepath.Dir(mdPath))
	tmplName := collection

	// Use the collection's template; default to "site.html" if not found
	tmpl := utils.Templates.Lookup(tmplName + ".html")
	if tmpl == nil {
		log.Printf("Template %s.html not found, using default site.html", tmplName)
		tmpl = utils.Templates.Lookup("site.html")
	}

	newdata, err := utils.LoadData(cfg.DataPath)
	if err != nil {
		log.Printf("Failed to load data: %v", err)
	}

	log.Printf("Executing template with Page: %+v", page)

	templateData := struct {
		Page *models.Content
		Data map[string]interface{}
	}{
		Page: page,
		Data: newdata,
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		return err
	}
	defer outputFile.Close()

	err = tmpl.ExecuteTemplate(outputFile, tmplName, templateData)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		return err
	}
	// Save the rendered HTML
	// err = os.WriteFile(outputPath, templateData.Page.Body, 0644)
	// if err != nil {
	// 	log.Printf("Error writing file: %v", err)
	// 	return err
	// }

	return nil
}

package ogimage

import (
	"log"
	"os"

	"os/exec"
	"path/filepath"
	"strings"

	"go-forth2.0/build/internal/models"
	"go-forth2.0/build/internal/utils"
)

// ContentItem represents a single item of content including its file path
type ContentItem struct {
	models.Content
	FilePath string
}

// GenerateAllOGImages walks through the content directory and generates OG images for each Markdown file
func GenerateAllOGImages(contentDir, outputDir string) error {
	contentItems, err := GetAllContentItems(contentDir)
	if err != nil {
		log.Printf("Error getting content items: %v", err)
		return err
	}

	for _, item := range contentItems {
		log.Printf("Processing file: %s", item.FilePath)
		if err := GenerateOGImageForFile(item.FilePath, outputDir); err != nil {
			log.Printf("Failed to generate OG image for %s: %v", item.FilePath, err)
		}
	}

	return nil
}

// GenerateOGImageForFile generates an OG image for a single Markdown file
func GenerateOGImageForFile(filePath, outputDir string) error {
	page, err := utils.LoadPageFromDirectory(filepath.Dir(filePath)+"/", filepath.Base(filePath))
	if err != nil {
		log.Printf("Error loading page: %v", err)
		return err
	}

	// Define paths for the temporary HTML file and the final OG image
	tempHTMLFilePath := filepath.Join(outputDir, "temp-og-image.html")
	imageName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)) + "-og-image.png"
	imagePath := filepath.Join(outputDir, imageName)

	if _, err := os.Stat(imagePath); err == nil {
		log.Printf("OG image already exists for %s, skipping...", filePath)
		return nil
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(filepath.Dir(imagePath), os.ModePerm); err != nil {
		log.Printf("Error creating output directory: %v", err)
		return err
	}

	// Render the OG image template to the temporary HTML file
	if err := utils.RenderTemplateStatic(tempHTMLFilePath, "og-image", page); err != nil {
		log.Printf("Error rendering OG image template: %v", err)
		return err
	}

	// Execute screenshot command
	cmd := exec.Command("node", "scripts/screenshot.js", tempHTMLFilePath, imagePath)
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to generate OG image for %s: %v", filePath, err)
		// Attempt to delete the temporary HTML file even if screenshot fails
		_ = os.Remove(tempHTMLFilePath)
		return err
	}

	// Delete the temporary HTML file
	if err := os.Remove(tempHTMLFilePath); err != nil {
		log.Printf("Error removing temporary file: %v", err)
	}

	page.OGImageURL = "/public/og-image/" + imageName

	return nil
}

// GetAllContentItems retrieves all content items from the content directory
func GetAllContentItems(contentDir string) ([]ContentItem, error) {
	var contentItems []ContentItem

	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			contentItem, err := processFile(path)
			if err != nil {
				log.Printf("Error processing file %s: %v", path, err)
				return nil
			}

			contentItems = append(contentItems, contentItem)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return contentItems, nil
}

// processFile processes a single Markdown file and returns a ContentItem
func processFile(filePath string) (ContentItem, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return ContentItem{}, err
	}

	frontMatter, _, err := utils.ParseFrontMatter(fileContent)
	if err != nil {
		return ContentItem{}, err
	}

	var content models.Content
	if title, ok := frontMatter["title"].(string); ok {
		content.Title = title
	}
	if description, ok := frontMatter["description"].(string); ok {
		content.Description = description
	}
	if date, ok := frontMatter["date"].(string); ok {
		content.Date = date
	}

	content.Collection = filepath.Base(filepath.Dir(filePath))

	return ContentItem{
		Content:  content,
		FilePath: filePath,
	}, nil
}

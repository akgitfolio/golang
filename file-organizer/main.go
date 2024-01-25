package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Parse command-line flags
	dir := flag.String("dir", ".", "Directory to organize")
	mode := flag.String("mode", "type", "Mode of organization: type, date")
	flag.Parse()

	// Validate the directory
	if _, err := os.Stat(*dir); os.IsNotExist(err) {
		log.Fatalf("Directory %s does not exist", *dir)
	}

	// Organize files based on the specified mode
	switch *mode {
	case "type":
		organizeByType(*dir)
	case "date":
		organizeByDate(*dir)
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}

// organizeByType organizes files into subdirectories based on their file types
func organizeByType(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext == "" {
			ext = "unknown"
		} else {
			ext = ext[1:] // Remove the dot
		}

		targetDir := filepath.Join(dir, ext)
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			os.Mkdir(targetDir, 0755)
		}

		oldPath := filepath.Join(dir, file.Name())
		newPath := filepath.Join(targetDir, file.Name())
		err := os.Rename(oldPath, newPath)
		if err != nil {
			log.Printf("Failed to move %s: %v", file.Name(), err)
		}
	}

	fmt.Println("Files organized by type")
}

// organizeByDate organizes files into subdirectories based on their creation dates
func organizeByDate(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		modTime := file.ModTime()
		dateDir := modTime.Format("2006-01-02")
		targetDir := filepath.Join(dir, dateDir)
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			os.Mkdir(targetDir, 0755)
		}

		oldPath := filepath.Join(dir, file.Name())
		newPath := filepath.Join(targetDir, file.Name())
		err := os.Rename(oldPath, newPath)
		if err != nil {
			log.Printf("Failed to move %s: %v", file.Name(), err)
		}
	}

	fmt.Println("Files organized by date")
}

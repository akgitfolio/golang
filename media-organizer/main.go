package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
	"github.com/rwcarlsen/goexif/exif"
)

// organizeImages processes and organizes image files
func organizeImages(srcDir, destDir string) {
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(srcDir, file.Name())
		f, err := os.Open(filePath)
		if err != nil {
			log.Println("Failed to open file:", filePath, err)
			continue
		}

		x, err := exif.Decode(f)
		f.Close()
		if err != nil {
			log.Println("Failed to decode EXIF data:", filePath, err)
			continue
		}

		date, err := x.DateTime()
		if err != nil {
			log.Println("Failed to get date from EXIF:", filePath, err)
			continue
		}

		newFileName := fmt.Sprintf("%s.jpg", date.Format("2006-01-02_15-04-05"))
		newDir := filepath.Join(destDir, date.Format("2006/01/02"))
		newFilePath := filepath.Join(newDir, newFileName)

		err = os.MkdirAll(newDir, os.ModePerm)
		if err != nil {
			log.Println("Failed to create directory:", newDir, err)
			continue
		}

		err = copyFile(filePath, newFilePath)
		if err != nil {
			log.Println("Failed to copy file:", filePath, newFilePath, err)
			continue
		}
	}
}

// organizeMusic processes and organizes music files
func organizeMusic(srcDir, destDir string) {
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(srcDir, file.Name())
		f, err := os.Open(filePath)
		if err != nil {
			log.Println("Failed to open file:", filePath, err)
			continue
		}

		m, err := tag.ReadFrom(f)
		f.Close()
		if err != nil {
			log.Println("Failed to read ID3 tags:", filePath, err)
			continue
		}

		artist := sanitize(m.Artist())
		album := sanitize(m.Album())
		title := sanitize(m.Title())

		newFileName := fmt.Sprintf("%s - %s.mp3", artist, title)
		newDir := filepath.Join(destDir, artist, album)
		newFilePath := filepath.Join(newDir, newFileName)

		err = os.MkdirAll(newDir, os.ModePerm)
		if err != nil {
			log.Println("Failed to create directory:", newDir, err)
			continue
		}

		err = copyFile(filePath, newFilePath)
		if err != nil {
			log.Println("Failed to copy file:", filePath, newFilePath, err)
			continue
		}
	}
}

// organizeVideos processes and organizes video files (by creation date)
func organizeVideos(srcDir, destDir string) {
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(srcDir, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			log.Println("Failed to get file info:", filePath, err)
			continue
		}

		date := info.ModTime()
		newFileName := fmt.Sprintf("%s.mp4", date.Format("2006-01-02_15-04-05"))
		newDir := filepath.Join(destDir, date.Format("2006/01/02"))
		newFilePath := filepath.Join(newDir, newFileName)

		err = os.MkdirAll(newDir, os.ModePerm)
		if err != nil {
			log.Println("Failed to create directory:", newDir, err)
			continue
		}

		err = copyFile(filePath, newFilePath)
		if err != nil {
			log.Println("Failed to copy file:", filePath, newFilePath, err)
			continue
		}
	}
}

// sanitize removes invalid characters from file and directory names
func sanitize(name string) string {
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(`\/:*?"<>|`, r) {
			return -1
		}
		return r
	}, name)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return dstFile.Sync()
}

func main() {
	srcDir := "path/to/source/directory"
	destDir := "path/to/destination/directory"

	// Organize images
	organizeImages(srcDir, filepath.Join(destDir, "images"))

	// Organize music
	organizeMusic(srcDir, filepath.Join(destDir, "music"))

	// Organize videos
	organizeVideos(srcDir, filepath.Join(destDir, "videos"))

	fmt.Println("Media files organized successfully.")
}

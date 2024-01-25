package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

func mergePDFs(outputFile string, inputFiles []string) error {
	err := api.MergeCreateFile(inputFiles, outputFile, nil)
	if err != nil {
		return fmt.Errorf("error merging PDFs: %w", err)
	}
	return nil
}

func splitPDF(inputFile string, outputDir string) error {
	err := api.SplitFile(inputFile, outputDir, 1, nil)
	if err != nil {
		return fmt.Errorf("error splitting PDF: %w", err)
	}
	return nil
}

func extractPages(inputFile string, outputFile string, pages []string) error {
	err := api.ExtractPagesFile(inputFile, outputFile, pages, nil)
	if err != nil {
		return fmt.Errorf("error extracting pages: %w", err)
	}
	return nil
}

func addWatermark(inputFile string, outputFile string, watermarkText string) error {
	watermarkConf := pdfcpu.DefaultWatermarkConfig()
	watermarkConf.Mode = pdfcpu.WMText
	watermarkConf.TextString = watermarkText
	err := api.AddWatermarksFile(inputFile, outputFile, nil, watermarkConf, nil)
	if err != nil {
		return fmt.Errorf("error adding watermark: %w", err)
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pdfmanipulator <command> [<args>]")
		fmt.Println("Commands: merge, split, extract, watermark")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "merge":
		if len(os.Args) < 4 {
			fmt.Println("Usage: pdfmanipulator merge <outputFile> <inputFiles...>")
			os.Exit(1)
		}
		outputFile := os.Args[2]
		inputFiles := os.Args[3:]
		err := mergePDFs(outputFile, inputFiles)
		if err != nil {
			log.Fatalf("Failed to merge PDFs: %v", err)
		}
		fmt.Println("PDFs merged successfully")

	case "split":
		if len(os.Args) != 4 {
			fmt.Println("Usage: pdfmanipulator split <inputFile> <outputDir>")
			os.Exit(1)
		}
		inputFile := os.Args[2]
		outputDir := os.Args[3]
		err := splitPDF(inputFile, outputDir)
		if err != nil {
			log.Fatalf("Failed to split PDF: %v", err)
		}
		fmt.Println("PDF split successfully")

	case "extract":
		if len(os.Args) < 5 {
			fmt.Println("Usage: pdfmanipulator extract <inputFile> <outputFile> <pages>")
			os.Exit(1)
		}
		inputFile := os.Args[2]
		outputFile := os.Args[3]
		pages := strings.Split(os.Args[4], ",")
		err := extractPages(inputFile, outputFile, pages)
		if err != nil {
			log.Fatalf("Failed to extract pages: %v", err)
		}
		fmt.Println("Pages extracted successfully")

	case "watermark":
		if len(os.Args) != 5 {
			fmt.Println("Usage: pdfmanipulator watermark <inputFile> <outputFile> <watermarkText>")
			os.Exit(1)
		}
		inputFile := os.Args[2]
		outputFile := os.Args[3]
		watermarkText := os.Args[4]
		err := addWatermark(inputFile, outputFile, watermarkText)
		if err != nil {
			log.Fatalf("Failed to add watermark: %v", err)
		}
		fmt.Println("Watermark added successfully")

	default:
		fmt.Println("Unknown command:", command)
		fmt.Println("Commands: merge, split, extract, watermark")
		os.Exit(1)
	}
}

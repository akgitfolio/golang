package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"

	"gopkg.in/yaml.v2"
)

func main() {
	inputFile := flag.String("input", "", "Input file path")
	outputFile := flag.String("output", "", "Output file path")
	format := flag.String("format", "csv", "Output format: csv, json, yaml")
	operation := flag.String("operation", "", "Operation: filter, sort, calculate")
	column := flag.String("column", "", "Column to operate on")
	value := flag.String("value", "", "Value to filter by or calculate")
	flag.Parse()

	if *inputFile == "" || *outputFile == "" || *operation == "" {
		log.Fatalf("Input file, output file, and operation are required")
	}

	data, err := readCSV(*inputFile)
	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}

	switch *operation {
	case "filter":
		if *column == "" || *value == "" {
			log.Fatalf("Column and value are required for filter operation")
		}
		data = filterRows(data, *column, *value)
	case "sort":
		if *column == "" {
			log.Fatalf("Column is required for sort operation")
		}
		data = sortRows(data, *column)
	case "calculate":
		if *column == "" || *value == "" {
			log.Fatalf("Column and value are required for calculate operation")
		}
		result := calculateColumn(data, *column, *value)
		fmt.Printf("Calculation result for column %s: %v\n", *column, result)
	default:
		log.Fatalf("Unknown operation: %s", *operation)
	}

	err = writeOutput(data, *outputFile, *format)
	if err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}
}

func readCSV(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

func filterRows(data [][]string, column, value string) [][]string {
	header := data[0]
	colIndex := -1
	for i, col := range header {
		if col == column {
			colIndex = i
			break
		}
	}
	if colIndex == -1 {
		log.Fatalf("Column %s not found", column)
	}

	filtered := [][]string{header}
	for _, row := range data[1:] {
		if row[colIndex] == value {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func sortRows(data [][]string, column string) [][]string {
	header := data[0]
	colIndex := -1
	for i, col := range header {
		if col == column {
			colIndex = i
			break
		}
	}
	if colIndex == -1 {
		log.Fatalf("Column %s not found", column)
	}

	rows := data[1:]
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][colIndex] < rows[j][colIndex]
	})

	sorted := append([][]string{header}, rows...)
	return sorted
}

func calculateColumn(data [][]string, column, operation string) float64 {
	header := data[0]
	colIndex := -1
	for i, col := range header {
		if col == column {
			colIndex = i
			break
		}
	}
	if colIndex == -1 {
		log.Fatalf("Column %s not found", column)
	}

	var result float64
	switch operation {
	case "sum":
		for _, row := range data[1:] {
			value, err := strconv.ParseFloat(row[colIndex], 64)
			if err != nil {
				log.Fatalf("Failed to parse value %s: %v", row[colIndex], err)
			}
			result += value
		}
	case "avg":
		count := 0
		for _, row := range data[1:] {
			value, err := strconv.ParseFloat(row[colIndex], 64)
			if err != nil {
				log.Fatalf("Failed to parse value %s: %v", row[colIndex], err)
			}
			result += value
			count++
		}
		result /= float64(count)
	default:
		log.Fatalf("Unknown calculation operation: %s", operation)
	}
	return result
}

func writeOutput(data [][]string, filePath, format string) error {
	switch format {
	case "csv":
		return writeCSV(data, filePath)
	case "json":
		return writeJSON(data, filePath)
	case "yaml":
		return writeYAML(data, filePath)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func writeCSV(data [][]string, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, record := range data {
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(data [][]string, filePath string) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, jsonData, 0644)
}

func writeYAML(data [][]string, filePath string) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, yamlData, 0644)
}

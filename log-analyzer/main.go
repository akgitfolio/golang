package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
}

var (
	logPattern = regexp.MustCompile(`\[(?P<timestamp>[^\]]+)\] \[(?P<level>[^\]]+)\] (?P<message>.*)`)
)

func main() {
	inputFile := flag.String("input", "", "Input log file path")
	outputFile := flag.String("output", "", "Output report file path")
	level := flag.String("level", "", "Log level to filter by (INFO, WARNING, ERROR)")
	flag.Parse()

	if *inputFile == "" {
		log.Fatalf("Input file is required")
	}

	logEntries, err := parseLogFile(*inputFile)
	if err != nil {
		log.Fatalf("Failed to parse log file: %v", err)
	}

	if *level != "" {
		logEntries = filterByLevel(logEntries, *level)
	}

	report := generateReport(logEntries)

	if *outputFile != "" {
		err = writeReport(*outputFile, report)
		if err != nil {
			log.Fatalf("Failed to write report: %v", err)
		}
	} else {
		fmt.Println(report)
	}
}

func parseLogFile(filePath string) ([]LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []LogEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		entry, err := parseLogLine(line)
		if err != nil {
			log.Printf("Failed to parse log line: %v", err)
			continue
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func parseLogLine(line string) (LogEntry, error) {
	matches := logPattern.FindStringSubmatch(line)
	if matches == nil {
		return LogEntry{}, fmt.Errorf("invalid log format")
	}

	timestampStr := matches[1]
	level := matches[2]
	message := matches[3]

	timestamp, err := time.Parse("2006-01-02 15:04:05", timestampStr)
	if err != nil {
		return LogEntry{}, err
	}

	return LogEntry{
		Timestamp: timestamp,
		Level:     level,
		Message:   message,
	}, nil
}

func filterByLevel(entries []LogEntry, level string) []LogEntry {
	var filtered []LogEntry
	for _, entry := range entries {
		if strings.EqualFold(entry.Level, level) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func generateReport(entries []LogEntry) string {
	var report strings.Builder
	report.WriteString("Log Report\n")
	report.WriteString("==========\n")
	report.WriteString(fmt.Sprintf("Total entries: %d\n", len(entries)))

	if len(entries) == 0 {
		return report.String()
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	startTime := entries[0].Timestamp
	endTime := entries[len(entries)-1].Timestamp
	report.WriteString(fmt.Sprintf("Time range: %s - %s\n", startTime, endTime))

	levelCount := make(map[string]int)
	for _, entry := range entries {
		levelCount[entry.Level]++
	}

	report.WriteString("Log levels:\n")
	for level, count := range levelCount {
		report.WriteString(fmt.Sprintf("  %s: %d\n", level, count))
	}

	report.WriteString("Entries:\n")
	for _, entry := range entries {
		report.WriteString(fmt.Sprintf("%s [%s] %s\n", entry.Timestamp, entry.Level, entry.Message))
	}

	return report.String()
}

func writeReport(filePath, report string) error {
	return os.WriteFile(filePath, []byte(report), 0644)
}

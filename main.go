package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
)

const inputFile = "input.txt"
const jsonFile = "settings.json"

var logger *slog.Logger

func main() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Application started")

	readInputFile(inputFile)
	readJsonSettings(jsonFile)

	logger.Info("Application ended")
}

func readInputFile(filename string) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Loop over each line and print it
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		logger.Error(fmt.Sprintf("Failed to open file: %v", err))
	} else {
		logger.Info("Read input file successfully")
	}
}

func readJsonSettings(filename string) {
	// Open the JSON settings file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Read the file contents
	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Decode into a map[string]bool because values are true/false
	settings := make(map[string]bool)
	if err := json.Unmarshal(data, &settings); err != nil {
		log.Fatalf("failed to parse JSON: %v", err)
	}

	// Print key/value pairs
	for k, v := range settings {
		fmt.Printf("%s = %v\n", k, v)
	}
	logger.Info("Read settings json successfully")
}

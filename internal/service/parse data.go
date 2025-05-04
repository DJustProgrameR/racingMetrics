package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

func ParseFiles(jsonConfigPath, eventsPath string) {
	jsonData, err := os.ReadFile(jsonConfigPath)
	if err != nil {
		log.Fatalf("Error reading JSON file: %v", err)
	}

	var config Pet
	err = json.Unmarshal(jsonData, &pet)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	fmt.Printf("Parsed JSON data: %+v\n", pet)

	file, err := os.Open(eventsPath)
	if err != nil {
		log.Fatalf("Error opening text file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines [][]string

	for scanner.Scan() {
		line := scanner.Text()
		args := strings.Fields(line)
		lines = append(lines, args)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading text file: %v", err)
	}

	fmt.Printf("Parsed lines from text file:\n")
	for _, args := range lines {
		fmt.Printf("%v\n", args)
	}
	err = json.Unmarshal(jsonData, &pet)
}

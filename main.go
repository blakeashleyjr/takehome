package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// FileInfo is a struct that holds the details of each file
type FileInfo struct {
	Name string
	Size int64
	Type string
	Path string
}

// log is a global logger that is faster and more useful than the standard logger
var log *zap.SugaredLogger

// CLI flags
var verbose bool
var index bool
var searchQuery string
var directory string

func init() {
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.BoolVar(&index, "i", false, "index files")
	flag.BoolVar(&index, "index", false, "index files")
	flag.StringVar(&searchQuery, "s", "", "search query")
	flag.StringVar(&searchQuery, "search", "", "search query")
	flag.StringVar(&directory, "d", "", "relative path to the directory to search")
	flag.StringVar(&directory, "directory", "", "relative path to the directory to search")
	flag.Parse()

	// If verbose flag is set, create a logger with debug level.
	// Otherwise, create a logger with info level.
	var logger *zap.Logger
	var err error

	if verbose {
		cfg := zap.NewDevelopmentConfig()
		cfg.Level.SetLevel(zap.DebugLevel)
		logger, err = cfg.Build()
		if err != nil {
			panic(err)
		}
	} else {
		cfg := zap.NewProductionConfig()
		cfg.Level.SetLevel(zap.InfoLevel)
		logger, err = cfg.Build()
		if err != nil {
			panic(err)
		}
	}

	log = logger.Sugar()
}

func main() {

	// If the directory flag is not provided but the index flag is, return an error
	if directory == "" && index {
		log.Fatalw("No directory flag provided. Please provide a relative path to the directory to index with the directory flag.")
	}

	// If both searchQuery and index are false, return an error
	if searchQuery == "" && !index {
		log.Fatalw("No search query or index flag provided. Please provide a search query and/or the index flag.")
	}

	// If search query is provided and index is not, run the search and exit
	if searchQuery != "" && !index {
		search(searchQuery)
		return
	}

	// Otherwise, index the files and exit

	// Create a slice to hold all the file information
	var files []FileInfo

	// Walk through the specified directory recursively
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Errorw("Error encountered while walking through files. Are you sure the directory exists and is correct?",
				"error", err,
			)
			return err
		}

		// Exclude any ".git" directory
		if strings.HasPrefix(info.Name(), ".git") {
			if info.IsDir() {
				return filepath.SkipDir // Skip the directory and all its subdirectories
			} else {
				return nil // Skip the file
			}
		}

		// If it's not a directory, it's a file
		if !info.IsDir() {
			// Open the file
			file, err := os.Open(path)
			if err != nil {
				log.Errorw("Error encountered while opening file",
					"file", path,
					"error", err,
				)
				return err
			}
			defer file.Close()

			// Create a buffer to read the content of the file
			buffer := make([]byte, 512)

			// Read from the file to the buffer
			_, err = file.Read(buffer)
			if err != nil {
				log.Errorw("Error encountered while reading file",
					"file", path,
					"error", err,
				)
				return err
			}

			// Attempt to detect the content type of the file
			contentType := http.DetectContentType(buffer)

			// Append the file details to the slice
			files = append(files, FileInfo{
				Name: info.Name(),
				Size: info.Size(),
				Type: contentType,
				Path: path,
			})

			// Log the file details
			log.Debugw("Successfully indexed file",
				"file", path,
				"name", info.Name(),
				"size", info.Size(),
				"type", contentType,
			)
		}
		return nil
	})

	// If an error occurred during the walk, log it
	if err != nil {
		log.Fatalw("Error encountered while walking through files",
			"error", err,
		)
	}

	// Create a new CSV file called "index.csv"
	file, err := os.Create("./index.csv")
	if err != nil {
		log.Fatalw("Error encountered while creating index.csv",
			"error", err,
		)
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the headers to the CSV file
	writer.Write([]string{"Name", "Size", "Type", "Path"})

	// Write each file's details as a row in the CSV file
	for _, fileInfo := range files {
		writer.Write([]string{
			fileInfo.Name,
			strconv.FormatInt(fileInfo.Size, 10),
			fileInfo.Type,
			fileInfo.Path,
		})
	}

	// Flush the data to the file
	writer.Flush()

	// Check if any error occurred while flushing
	if err := writer.Error(); err != nil {
		log.Fatalw("Error encountered while writing to index.csv",
			"error", err,
		)
	}

	// Log the creation of the index file
	log.Infow("Successfully created index file",
		"filename", "index.csv",
		"fileCount", len(files),
	)

	// If the search query and the index flag are provided, run the search
	if searchQuery != "" && index {
		search(searchQuery)
		return
	}
}

func search(query string) {

	// Open the index file
	file, err := os.Open("./index.csv")
	if err != nil {
		log.Fatalw("Failed to open index file or the file does not exist. Be sure to run the program with the -i flag to create an index.csv file", "error", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()

	// Check if the lines slice is empty
	if len(lines) == 0 {
		log.Warnw("Index file is empty.")
		return
	}

	if err != nil {
		log.Fatalw("Failed to read index file", "error", err)
	}

	// The first line is the header, skip it
	for _, line := range lines[1:] {
		// Make sure the line has at least one column. Ran into "slice bounds out of range" error without this check
		if len(line) > 0 {
			// We assume that Name is in the first column
			if strings.Contains(line[0], query) {
				fmt.Println(line)
			}
		}
	}
}

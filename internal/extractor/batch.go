package extractor

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// BatchProcessor handles processing multiple PDF files
type BatchProcessor struct {
	InputPattern string
	OutputDir    string
	Workers      int
	Verbose      bool
}

// ProcessResult holds the result of processing a single file
type ProcessResult struct {
	Filename   string
	OutputPath string
	Error      error
}

// ProcessBatch processes multiple PDF files in parallel
func (bp *BatchProcessor) ProcessBatch() error {
	// Find all PDF files matching the pattern
	files, err := filepath.Glob(bp.InputPattern)
	if err != nil {
		return fmt.Errorf("Fehler beim Suchen von Dateien: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("Keine PDF-Dateien gefunden, die dem Muster entsprechen: %s", bp.InputPattern)
	}

	// Nur PDF-Dateien filtern
	pdfFiles := make([]string, 0, len(files))
	for _, file := range files {
		if strings.ToLower(filepath.Ext(file)) == ".pdf" {
			pdfFiles = append(pdfFiles, file)
		}
	}

	if len(pdfFiles) == 0 {
		return fmt.Errorf("Keine PDF-Dateien gefunden, die dem Muster entsprechen: %s", bp.InputPattern)
	}

	fmt.Printf("Gefunden: %d PDF-Dateien zur Verarbeitung\n", len(pdfFiles))

	// Create worker pool
	jobs := make(chan string, len(pdfFiles))
	results := make(chan ProcessResult, len(pdfFiles))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < bp.Workers; i++ {
		wg.Add(1)
		go bp.worker(jobs, results, &wg)
	}

	// Send jobs
	for _, file := range pdfFiles {
		jobs <- file
	}
	close(jobs)

	// Collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results
	successful := 0
	failed := 0
	for result := range results {
		if result.Error != nil {
			fmt.Printf("❌ %s: %v\n", result.Filename, result.Error)
			failed++
		} else {
			fmt.Printf("✅ %s -> %s\n", result.Filename, result.OutputPath)
			successful++
		}
	}

	fmt.Printf("\nBatch-Verarbeitung abgeschlossen: %d erfolgreich, %d fehlgeschlagen\n", successful, failed)
	return nil
}

// worker processes files from the jobs channel
func (bp *BatchProcessor) worker(jobs <-chan string, results chan<- ProcessResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for filename := range jobs {
		// Bestimme Ausgabepfad
		var outputPath string
		if bp.OutputDir != "" {
			baseName := filepath.Base(filename)
			baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
			outputPath = filepath.Join(bp.OutputDir, baseName+".xml")
		}

		extractor := &ZUGFeRDExtractor{
			InputPath:  filename,
			OutputPath: outputPath,
			Verbose:    bp.Verbose,
		}

		err := extractor.ExtractXML()

		// Ermittle tatsächlichen Ausgabepfad für die Erfolgsbenachrichtigung
		result := ProcessResult{
			Filename: filename,
			Error:    err,
		}

		if err == nil {
			// Wenn kein expliziter Ausgabepfad angegeben wurde, nehmen wir den generierten von ExtractXML
			if outputPath == "" {
				baseName := filepath.Base(filename)
				baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
				result.OutputPath = filepath.Join(filepath.Dir(filename), baseName+".xml")
			} else {
				result.OutputPath = outputPath
			}
		}

		results <- result
	}
}

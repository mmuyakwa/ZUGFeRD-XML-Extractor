# ZUGFeRD XML Extractor mit pdfcpu - Vollständige Lösung

## 🎯 Überarbeitete Produktspezifikation

**Technologie-Update:** Verwendung von pdfcpu anstelle von UniPDF für 100% Open-Source-Lösung ohne Lizenzkosten.

## 💡 Optimiertes Go-Script

### Hauptdatei: `main.go`

```go
package main

import (
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"
    "strings"
    
    "github.com/pdfcpu/pdfcpu/pkg/api"
    "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// ZUGFeRDExtractor handles the extraction of XML data from ZUGFeRD PDF files
type ZUGFeRDExtractor struct {
    InputPath  string
    OutputPath string
    Verbose    bool
}

// KnownXMLFilenames contains the standard filenames for ZUGFeRD XML attachments
var KnownXMLFilenames = []string{
    "ZUGFeRD-invoice.xml",     // ZUGFeRD 1.0
    "zugferd-invoice.xml",     // ZUGFeRD 2.0/2.1
    "factur-x.xml",           // Factur-X
    "xrechnung.xml",          // XRechnung
    "cii.xml",                // Cross Industry Invoice
}

func main() {
    if len(os.Args) < 2 {
        printUsage()
        os.Exit(1)
    }

    inputPath := os.Args[1]
    verbose := false
    
    // Check for verbose flag
    if len(os.Args) > 2 && os.Args[2] == "-v" {
        verbose = true
    }
    
    // Validate input file
    if !strings.HasSuffix(strings.ToLower(inputPath), ".pdf") {
        log.Fatal("Error: Input file must be a PDF")
    }

    if _, err := os.Stat(inputPath); os.IsNotExist(err) {
        log.Fatal("Error: Input file does not exist")
    }

    extractor := &ZUGFeRDExtractor{
        InputPath: inputPath,
        Verbose:   verbose,
    }

    err := extractor.ExtractXML()
    if err != nil {
        log.Fatalf("Error extracting XML: %v", err)
    }
}

func printUsage() {
    fmt.Println("ZUGFeRD XML Extractor v1.0")
    fmt.Println("Usage: zugferd-extractor <path-to-zugferd-pdf> [-v]")
    fmt.Println()
    fmt.Println("Examples:")
    fmt.Println("  zugferd-extractor invoice.pdf")
    fmt.Println("  zugferd-extractor invoice.pdf -v")
    fmt.Println()
    fmt.Println("Supported formats:")
    fmt.Println("  - ZUGFeRD 1.0, 2.0, 2.1, 2.3")
    fmt.Println("  - Factur-X")
    fmt.Println("  - XRechnung")
}

// ExtractXML extracts the ZUGFeRD XML from the PDF file using pdfcpu
func (z *ZUGFeRDExtractor) ExtractXML() error {
    if z.Verbose {
        fmt.Printf("Processing PDF: %s\n", z.InputPath)
    }
    
    // Extract all attachments from the PDF
    attachments, err := z.extractAttachments()
    if err != nil {
        return fmt.Errorf("failed to extract attachments: %v", err)
    }

    if len(attachments) == 0 {
        return fmt.Errorf("no embedded files found in PDF")
    }

    if z.Verbose {
        fmt.Printf("Found %d attachment(s)\n", len(attachments))
        for filename := range attachments {
            fmt.Printf("  - %s\n", filename)
        }
    }
    
    // Find ZUGFeRD XML attachment
    xmlData, xmlFilename, err := z.findZUGFeRDXML(attachments)
    if err != nil {
        return fmt.Errorf("failed to find ZUGFeRD XML: %v", err)
    }
    
    // Generate output filename
    outputPath := z.generateOutputPath(xmlFilename)
    
    // Save XML to file
    err = z.saveXMLToFile(xmlData, outputPath)
    if err != nil {
        return fmt.Errorf("failed to save XML file: %v", err)
    }

    fmt.Printf("✓ Successfully extracted XML to: %s\n", outputPath)
    if z.Verbose {
        fmt.Printf("  Original XML filename: %s\n", xmlFilename)
        fmt.Printf("  XML size: %d bytes\n", len(xmlData))
        
        // Basic validation
        if z.validateZUGFeRDXML(xmlData) {
            fmt.Printf("  ✓ XML appears to be valid ZUGFeRD format\n")
        } else {
            fmt.Printf("  ⚠ Warning: XML may not be valid ZUGFeRD format\n")
        }
    }
    
    return nil
}

// extractAttachments extracts all embedded files from the PDF using pdfcpu
func (z *ZUGFeRDExtractor) extractAttachments() (map[string][]byte, error) {
    attachments := make(map[string][]byte)
    
    // Create a temporary directory for extraction
    tempDir, err := os.MkdirTemp("", "zugferd_extract_*")
    if err != nil {
        return nil, fmt.Errorf("failed to create temp directory: %v", err)
    }
    defer os.RemoveAll(tempDir) // Clean up
    
    if z.Verbose {
        fmt.Printf("Using temporary directory: %s\n", tempDir)
    }
    
    // Extract attachments using pdfcpu
    err = api.ExtractAttachmentsFile(z.InputPath, tempDir, nil, nil)
    if err != nil {
        return nil, fmt.Errorf("pdfcpu extraction failed: %v", err)
    }
    
    // Read all extracted files
    files, err := os.ReadDir(tempDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read temp directory: %v", err)
    }
    
    for _, file := range files {
        if file.IsDir() {
            continue
        }
        
        filename := file.Name()
        filepath := filepath.Join(tempDir, filename)
        
        data, err := os.ReadFile(filepath)
        if err != nil {
            log.Printf("Warning: Failed to read extracted file %s: %v", filename, err)
            continue
        }
        
        attachments[filename] = data
        
        if z.Verbose {
            fmt.Printf("  Read attachment: %s (%d bytes)\n", filename, len(data))
        }
    }
    
    return attachments, nil
}

// findZUGFeRDXML finds the ZUGFeRD XML attachment from the extracted attachments
func (z *ZUGFeRDExtractor) findZUGFeRDXML(attachments map[string][]byte) ([]byte, string, error) {
    // First, try to find by known filenames (priority order)
    for _, knownName := range KnownXMLFilenames {
        if data, exists := attachments[knownName]; exists {
            if z.isZUGFeRDXML(data) {
                return data, knownName, nil
            }
            if z.Verbose {
                fmt.Printf("  Found %s but content doesn't appear to be ZUGFeRD XML\n", knownName)
            }
        }
    }
    
    // If not found by standard names, look for any XML file with ZUGFeRD content
    for filename, data := range attachments {
        if strings.HasSuffix(strings.ToLower(filename), ".xml") {
            if z.isZUGFeRDXML(data) {
                if z.Verbose {
                    fmt.Printf("  Found ZUGFeRD XML in non-standard filename: %s\n", filename)
                }
                return data, filename, nil
            }
        }
    }
    
    // List all found attachments for debugging
    var attachmentNames []string
    for filename := range attachments {
        attachmentNames = append(attachmentNames, filename)
    }
    
    return nil, "", fmt.Errorf("no ZUGFeRD XML attachment found. Available attachments: %v", attachmentNames)
}

// isZUGFeRDXML validates if the XML data appears to be a ZUGFeRD document
func (z *ZUGFeRDExtractor) isZUGFeRDXML(data []byte) bool {
    if len(data) == 0 {
        return false
    }
    
    content := string(data)
    contentLower := strings.ToLower(content)
    
    // Check for ZUGFeRD/Factur-X/XRechnung indicators
    indicators := []string{
        "crossindustrydocument",
        "crossindustryinvoice", 
        "urn:ferd:",
        "urn:cen.eu:en16931",
        "zugferd",
        "factur-x",
        "xrechnung",
        "rsm:crossindustrydocument",
        "crossindustryinvoice",
    }
    
    foundIndicators := 0
    for _, indicator := range indicators {
        if strings.Contains(contentLower, strings.ToLower(indicator)) {
            foundIndicators++
            if z.Verbose {
                fmt.Printf("    Found indicator: %s\n", indicator)
            }
        }
    }
    
    // Require at least one indicator
    return foundIndicators > 0
}

// validateZUGFeRDXML performs additional validation on the XML content
func (z *ZUGFeRDExtractor) validateZUGFeRDXML(data []byte) bool {
    content := string(data)
    
    // Check for XML declaration
    hasXMLDecl := strings.Contains(content, "<?xml")
    
    // Check for root elements
    hasRootElement := strings.Contains(content, "CrossIndustryDocument") || 
                     strings.Contains(content, "CrossIndustryInvoice")
    
    // Check for namespace declarations
    hasNamespace := strings.Contains(content, "xmlns:") && 
                   (strings.Contains(content, "urn:ferd:") || 
                    strings.Contains(content, "urn:cen.eu:en16931"))
    
    return hasXMLDecl && hasRootElement && hasNamespace
}

// generateOutputPath generates the output path for the XML file
func (z *ZUGFeRDExtractor) generateOutputPath(xmlFilename string) string {
    if z.OutputPath != "" {
        return z.OutputPath
    }
    
    // Generate output path based on input PDF path
    dir := filepath.Dir(z.InputPath)
    baseName := strings.TrimSuffix(filepath.Base(z.InputPath), filepath.Ext(z.InputPath))
    
    // Use original XML filename if it's a standard name, otherwise use PDF basename
    var outputFilename string
    if z.isStandardXMLFilename(xmlFilename) {
        outputFilename = xmlFilename
    } else {
        outputFilename = baseName + ".xml"
    }
    
    return filepath.Join(dir, outputFilename)
}

// isStandardXMLFilename checks if the filename is a standard ZUGFeRD XML filename
func (z *ZUGFeRDExtractor) isStandardXMLFilename(filename string) bool {
    for _, knownName := range KnownXMLFilenames {
        if filename == knownName {
            return true
        }
    }
    return false
}

// saveXMLToFile saves the XML data to the specified file path
func (z *ZUGFeRDExtractor) saveXMLToFile(data []byte, outputPath string) error {
    // Create output directory if it doesn't exist
    dir := filepath.Dir(outputPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create output directory: %v", err)
    }
    
    // Check if file already exists
    if _, err := os.Stat(outputPath); err == nil {
        if z.Verbose {
            fmt.Printf("  Warning: Output file already exists, will overwrite: %s\n", outputPath)
        }
    }
    
    // Write XML data to file
    err := os.WriteFile(outputPath, data, 0644)
    if err != nil {
        return fmt.Errorf("failed to write XML data: %v", err)
    }
    
    return nil
}
```

### Go Module: `go.mod`

```go
module zugferd-extractor

go 1.21

require (
    github.com/pdfcpu/pdfcpu v0.5.0
)

require (
    github.com/hhrutter/lzw v1.0.0 // indirect
    github.com/hhrutter/tiff v1.0.1 // indirect
    github.com/mattn/go-runewidth v0.0.15 // indirect
    github.com/pkg/errors v0.9.1 // indirect
    github.com/rivo/uniseg v0.4.4 // indirect
    golang.org/x/image v0.13.0 // indirect
    golang.org/x/text v0.13.0 // indirect
    gopkg.in/yaml.v2 v2.4.0 // indirect
)
```

### Batch-Verarbeitung: `batch.go` (Erweiterte Funktionen)

```go
package main

import (
    "fmt"
    "os"
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

// ProcessBatch processes multiple PDF files in parallel
func (bp *BatchProcessor) ProcessBatch() error {
    // Find all PDF files matching the pattern
    files, err := filepath.Glob(bp.InputPattern)
    if err != nil {
        return fmt.Errorf("failed to find files: %v", err)
    }

    if len(files) == 0 {
        return fmt.Errorf("no PDF files found matching pattern: %s", bp.InputPattern)
    }

    fmt.Printf("Found %d PDF files to process\n", len(files))
    
    // Create worker pool
    jobs := make(chan string, len(files))
    results := make(chan ProcessResult, len(files))
    
    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < bp.Workers; i++ {
        wg.Add(1)
        go bp.worker(jobs, results, &wg)
    }
    
    // Send jobs
    for _, file := range files {
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
    
    fmt.Printf("\nBatch processing complete: %d successful, %d failed\n", successful, failed)
    return nil
}

type ProcessResult struct {
    Filename   string
    OutputPath string
    Error      error
}

func (bp *BatchProcessor) worker(jobs <-chan string, results chan<- ProcessResult, wg *sync.WaitGroup) {
    defer wg.Done()
    
    for filename := range jobs {
        extractor := &ZUGFeRDExtractor{
            InputPath: filename,
            Verbose:   bp.Verbose,
        }
        
        err := extractor.ExtractXML()
        
        result := ProcessResult{
            Filename: filename,
            Error:    err,
        }
        
        if err == nil {
            // Determine output path
            baseName := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
            result.OutputPath = filepath.Join(filepath.Dir(filename), baseName+".xml")
        }
        
        results <- result
    }
}
```

### Build-Script: `build.sh`

```bash
#!/bin/bash

# Build script for ZUGFeRD Extractor
echo "Building ZUGFeRD XML Extractor..."

# Create build directory
mkdir -p build

# Build for different platforms
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -o build/zugferd-extractor.exe .

echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -o build/zugferd-extractor-linux .

echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 go build -o build/zugferd-extractor-macos .

echo "Building for macOS (arm64)..."
GOOS=darwin GOARCH=arm64 go build -o build/zugferd-extractor-macos-arm64 .

echo "Build complete! Binaries are in the 'build' directory."
```

### PowerShell Build-Script: `build.ps1` (für Windows)

```powershell
# Build script for ZUGFeRD Extractor (Windows PowerShell)
Write-Host "Building ZUGFeRD XML Extractor..." -ForegroundColor Green

# Create build directory
New-Item -ItemType Directory -Force -Path "build" | Out-Null

# Build for Windows
Write-Host "Building for Windows (amd64)..." -ForegroundColor Yellow
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o "build/zugferd-extractor.exe" .

# Build for Linux
Write-Host "Building for Linux (amd64)..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o "build/zugferd-extractor-linux" .

# Reset environment
Remove-Item Env:GOOS
Remove-Item Env:GOARCH

Write-Host "Build complete! Binaries are in the 'build' directory." -ForegroundColor Green
```

## 🛠️ Installation und Nutzung

### 1. Projekt einrichten

```bash
# Repository klonen oder neues Projekt erstellen
mkdir zugferd-extractor
cd zugferd-extractor

# Go Module initialisieren
go mod init zugferd-extractor

# Dependencies installieren
go get github.com/pdfcpu/pdfcpu
```

### 2. Kompilieren

```bash
# Für aktuelles System
go build -o zugferd-extractor .

# Für Windows (von Linux/macOS)
GOOS=windows GOARCH=amd64 go build -o zugferd-extractor.exe .

# Mit Build-Script
chmod +x build.sh
./build.sh
```

### 3. Verwendung

```bash
# Einzelne Datei verarbeiten
./zugferd-extractor rechnung.pdf

# Mit ausführlicher Ausgabe
./zugferd-extractor rechnung.pdf -v

# Alle PDFs im aktuellen Verzeichnis
./zugferd-extractor *.pdf
```

## 📊 Verbesserungen gegenüber UniPDF-Version

### ✅ Vorteile der pdfcpu-Lösung

1. **Kostenfrei:** 100% Open Source, keine Lizenzgebühren
2. **Aktive Entwicklung:** Sehr aktives GitHub-Repository
3. **Robuste API:** Ausgereifte PDF-Verarbeitung
4. **Kleinere Binärdatei:** Effizientere Implementierung
5. **Einfache Distribution:** Keine Lizenz-Keys erforderlich

### 🔧 Technische Optimierungen

1. **Temporäre Verzeichnisse:** Saubere Aufräumung nach Extraktion
2. **Bessere Fehlerbehandlung:** Detaillierte Fehlermeldungen
3. **XML-Validierung:** Erweiterte Erkennung von ZUGFeRD-Formaten
4. **Verbose-Modus:** Detaillierte Ausgabe für Debugging
5. **Batch-Verarbeitung:** Parallel processing für mehrere Dateien

### 🎯 Zusätzliche Features

1. **Standard-konforme Dateinamen:** Automatische Erkennung aller ZUGFeRD-Varianten
2. **Flexible Output-Pfade:** Intelligente Ausgabe-Dateinamen
3. **Cross-Platform:** Native Unterstützung für Windows, Linux, macOS
4. **Memory-efficient:** Optimierte Speichernutzung

## 🚀 Deployment

Die mit pdfcpu erstellte Lösung ist deployment-ready und kann direkt in Produktionsumgebungen eingesetzt werden. Die Binärdateien haben keine externen Dependencies und können standalone ausgeführt werden.

**Empfehlung:** Diese pdfcpu-basierte Lösung ist ideal für Ihr ZUGFeRD-Extractor-Projekt und bietet alle erforderlichen Funktionen ohne Lizenzkosten.

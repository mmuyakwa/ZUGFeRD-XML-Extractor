package extractor

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// ZUGFeRDExtractor handles the extraction of XML data from ZUGFeRD PDF files
type ZUGFeRDExtractor struct {
	InputPath  string
	OutputPath string
	Verbose    bool
}

// KnownXMLFilenames contains the standard filenames for ZUGFeRD XML attachments
var KnownXMLFilenames = []string{
	"ZUGFeRD-invoice.xml", // ZUGFeRD 1.0
	"zugferd-invoice.xml", // ZUGFeRD 2.0/2.1
	"factur-x.xml",        // Factur-X
	"xrechnung.xml",       // XRechnung
	"cii.xml",             // Cross Industry Invoice
}

// ExtractXML extracts the ZUGFeRD XML from the PDF file using multiple approaches
func (z *ZUGFeRDExtractor) ExtractXML() error {
	if z.Verbose {
		fmt.Printf("Verarbeite PDF: %s\n", z.InputPath)
	}

	// Try multiple extraction methods
	var attachments map[string][]byte
	var err error

	// Method 1: Try standard pdfcpu extraction
	attachments, err = z.extractAttachmentsStandard()
	if err != nil {
		if z.Verbose {
			fmt.Printf("Standard-Extraktion fehlgeschlagen: %v\n", err)
			fmt.Printf("Versuche relaxierte Extraktion...\n")
		}

		// Method 2: Try with relaxed validation
		attachments, err = z.extractAttachmentsRelaxed()
		if err != nil {
			if z.Verbose {
				fmt.Printf("Relaxierte Extraktion fehlgeschlagen: %v\n", err)
				fmt.Printf("Versuche manuelle Extraktion...\n")
			}

			// Method 3: Try manual extraction
			attachments, err = z.extractAttachmentsManual()
			if err != nil {
				return fmt.Errorf("alle Extraktionsmethoden fehlgeschlagen: %v", err)
			}
		}
	}

	if len(attachments) == 0 {
		return fmt.Errorf("keine eingebetteten Dateien im PDF gefunden")
	}

	if z.Verbose {
		fmt.Printf("Gefunden: %d Anhang/Anhänge\n", len(attachments))
		for filename := range attachments {
			fmt.Printf("  - %s\n", filename)
		}
	}

	// Find ZUGFeRD XML attachment
	xmlData, xmlFilename, err := z.findZUGFeRDXML(attachments)
	if err != nil {
		return fmt.Errorf("ZUGFeRD XML nicht gefunden: %v", err)
	}

	// Generate output filename
	outputPath := z.generateOutputPath(xmlFilename)

	// Save XML to file
	err = z.saveXMLToFile(xmlData, outputPath)
	if err != nil {
		return fmt.Errorf("Fehler beim Speichern der XML-Datei: %v", err)
	}

	fmt.Printf("✓ XML erfolgreich extrahiert nach: %s\n", outputPath)
	if z.Verbose {
		fmt.Printf("  Originaler XML-Dateiname: %s\n", xmlFilename)
		fmt.Printf("  XML-Größe: %d Bytes\n", len(xmlData))

		// Basic validation
		if z.validateZUGFeRDXML(xmlData) {
			fmt.Printf("  ✓ XML scheint ein gültiges ZUGFeRD-Format zu sein\n")
		} else {
			fmt.Printf("  ⚠ Warnung: XML könnte kein gültiges ZUGFeRD-Format sein\n")
		}
	}

	return nil
}

// extractAttachmentsStandard tries standard pdfcpu extraction
func (z *ZUGFeRDExtractor) extractAttachmentsStandard() (map[string][]byte, error) {
	attachments := make(map[string][]byte)

	// Create a temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "zugferd_extract_*")
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Erstellen des temporären Verzeichnisses: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if z.Verbose {
		fmt.Printf("Verwende temporäres Verzeichnis: %s\n", tempDir)
	}

	// Extract attachments using pdfcpu
	err = api.ExtractAttachmentsFile(z.InputPath, tempDir, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("pdfcpu-Extraktion fehlgeschlagen: %v", err)
	}

	return z.readExtractedFiles(tempDir)
}

// extractAttachmentsRelaxed tries extraction with relaxed validation
func (z *ZUGFeRDExtractor) extractAttachmentsRelaxed() (map[string][]byte, error) {
	attachments := make(map[string][]byte)

	tempDir, err := os.MkdirTemp("", "zugferd_extract_relaxed_*")
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Erstellen des temporären Verzeichnisses: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create relaxed configuration
	config := model.NewDefaultConfiguration()
	config.ValidationMode = model.ValidationRelaxed
	config.DecodeAllStreams = false

	err = api.ExtractAttachmentsFile(z.InputPath, tempDir, nil, config)
	if err != nil {
		return nil, fmt.Errorf("relaxierte pdfcpu-Extraktion fehlgeschlagen: %v", err)
	}

	return z.readExtractedFiles(tempDir)
}

// extractAttachmentsManual tries manual extraction by parsing PDF structure
func (z *ZUGFeRDExtractor) extractAttachmentsManual() (map[string][]byte, error) {
	// This is a simplified manual extraction - in practice you'd need more robust PDF parsing
	file, err := os.Open(z.InputPath)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Öffnen der PDF: %v", err)
	}
	defer file.Close()

	// Read the entire file
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Lesen der PDF: %v", err)
	}

	attachments := make(map[string][]byte)

	// Look for embedded file markers and XML content
	content := string(data)

	// Search for common ZUGFeRD XML patterns
	xmlStartPatterns := []string{
		"<?xml version=\"1.0\"",
		"<rsm:CrossIndustryDocument",
		"<rsm:CrossIndustryInvoice",
	}

	for _, pattern := range xmlStartPatterns {
		if idx := strings.Index(content, pattern); idx != -1 {
			// Try to extract XML from this position
			xmlData := z.extractXMLFromPosition(data, idx)
			if len(xmlData) > 0 && z.isZUGFeRDXML(xmlData) {
				// Guess filename based on content
				filename := z.guessXMLFilename(xmlData)
				attachments[filename] = xmlData
				if z.Verbose {
					fmt.Printf("  XML manuell extrahiert von Position %d\n", idx)
				}
			}
		}
	}

	if len(attachments) == 0 {
		return nil, fmt.Errorf("manuelle Extraktion fand keine XML-Anhänge")
	}

	return attachments, nil
}

// extractXMLFromPosition attempts to extract XML content starting from a position
func (z *ZUGFeRDExtractor) extractXMLFromPosition(data []byte, startPos int) []byte {
	if startPos < 0 || startPos >= len(data) {
		return nil
	}

	content := string(data[startPos:])

	// Find the end of the XML document
	endPatterns := []string{
		"</rsm:CrossIndustryDocument>",
		"</rsm:CrossIndustryInvoice>",
		"</CrossIndustryDocument>",
		"</CrossIndustryInvoice>",
	}

	maxLength := len(content)
	endPos := maxLength

	for _, pattern := range endPatterns {
		if idx := strings.Index(content, pattern); idx != -1 {
			candidateEnd := idx + len(pattern)
			if candidateEnd < endPos {
				endPos = candidateEnd
			}
		}
	}

	if endPos < maxLength {
		xmlContent := content[:endPos]
		return []byte(xmlContent)
	}

	return nil
}

// guessXMLFilename guesses the appropriate filename based on XML content
func (z *ZUGFeRDExtractor) guessXMLFilename(data []byte) string {
	content := strings.ToLower(string(data))

	if strings.Contains(content, "xrechnung") {
		return "xrechnung.xml"
	} else if strings.Contains(content, "factur-x") {
		return "factur-x.xml"
	} else if strings.Contains(content, "zugferd") {
		if strings.Contains(content, "urn:ferd:pdfa:crossindustrydocument:invoice:1p0") {
			return "ZUGFeRD-invoice.xml" // Version 1.0
		} else {
			return "zugferd-invoice.xml" // Version 2.0+
		}
	}

	return "invoice.xml" // fallback
}

// readExtractedFiles reads all files from the extraction directory
func (z *ZUGFeRDExtractor) readExtractedFiles(tempDir string) (map[string][]byte, error) {
	attachments := make(map[string][]byte)

	files, err := os.ReadDir(tempDir)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Lesen des temporären Verzeichnisses: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		filepath := filepath.Join(tempDir, filename)

		data, err := os.ReadFile(filepath)
		if err != nil {
			log.Printf("Warnung: Konnte Datei nicht lesen %s: %v", filename, err)
			continue
		}

		attachments[filename] = data

		if z.Verbose {
			fmt.Printf("  Anhang gelesen: %s (%d Bytes)\n", filename, len(data))
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
				if z.Verbose {
					fmt.Printf("  Standard-ZUGFeRD-XML gefunden: %s\n", knownName)
				}
				return data, knownName, nil
			}
			if z.Verbose {
				fmt.Printf("  %s gefunden, aber Inhalt scheint keine ZUGFeRD-XML zu sein\n", knownName)
			}
		}
	}

	// If not found by standard names, look for any XML file with ZUGFeRD content
	for filename, data := range attachments {
		if strings.HasSuffix(strings.ToLower(filename), ".xml") {
			if z.isZUGFeRDXML(data) {
				if z.Verbose {
					fmt.Printf("  ZUGFeRD-XML in nicht-standardisiertem Dateinamen gefunden: %s\n", filename)
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

	return nil, "", fmt.Errorf("kein ZUGFeRD-XML-Anhang gefunden. Verfügbare Anhänge: %v", attachmentNames)
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
				fmt.Printf("    Indikator gefunden: %s\n", indicator)
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
		return fmt.Errorf("Fehler beim Erstellen des Ausgabeverzeichnisses: %v", err)
	}

	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		if z.Verbose {
			fmt.Printf("  Warnung: Ausgabedatei existiert bereits und wird überschrieben: %s\n", outputPath)
		}
	}

	// Write XML data to file
	err := os.WriteFile(outputPath, data, 0644)
	if err != nil {
		return fmt.Errorf("Fehler beim Schreiben der XML-Daten: %v", err)
	}

	return nil
}

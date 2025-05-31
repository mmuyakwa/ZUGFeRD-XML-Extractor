package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"zugferd-extractor/internal/extractor"
)

func main() {
	// Kommandozeilenargumente definieren
	verbosePtr := flag.Bool("v", false, "Ausführliche Ausgabe")
	outputPtr := flag.String("o", "", "Ausgabepfad für die XML-Datei")
	helpPtr := flag.Bool("h", false, "Hilfe anzeigen")
	flag.Parse()

	// Hilfe anzeigen, wenn angefordert oder keine Argumente vorhanden
	if *helpPtr || flag.NArg() < 1 {
		printUsage()
		if *helpPtr {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	// Extractor konfigurieren
	inputPattern := flag.Arg(0)
	verbose := *verbosePtr
	outputPath := *outputPtr

	// Prüfen, ob Batch-Verarbeitung oder einzelne Datei
	files, err := filepath.Glob(inputPattern)
	if err != nil {
		log.Fatalf("Fehler beim Suchen von Dateien: %v", err)
	}

	// Keine übereinstimmenden Dateien gefunden
	if len(files) == 0 {
		log.Fatalf("Keine Dateien gefunden, die dem Muster '%s' entsprechen", inputPattern)
	}

	// Batchverarbeitung für mehrere Dateien
	if len(files) > 1 {
		// Wenn ein Ausgabepfad angegeben wurde, muss es ein Verzeichnis sein
		if outputPath != "" {
			info, err := os.Stat(outputPath)
			if err != nil {
				if os.IsNotExist(err) {
					err = os.MkdirAll(outputPath, 0755)
					if err != nil {
						log.Fatalf("Fehler beim Erstellen des Ausgabeverzeichnisses: %v", err)
					}
				} else {
					log.Fatalf("Fehler beim Überprüfen des Ausgabepfads: %v", err)
				}
			} else if !info.IsDir() {
				log.Fatalf("Ausgabepfad muss ein Verzeichnis sein, wenn mehrere Dateien verarbeitet werden")
			}
		}

		// Anzahl der Worker basierend auf CPU-Kernen
		numWorkers := runtime.NumCPU()
		if numWorkers > len(files) {
			numWorkers = len(files)
		}

		processor := &extractor.BatchProcessor{
			InputPattern: inputPattern,
			OutputDir:    outputPath,
			Workers:      numWorkers,
			Verbose:      verbose,
		}

		if err := processor.ProcessBatch(); err != nil {
			log.Fatalf("Batch-Verarbeitungsfehler: %v", err)
		}
		return
	}

	// Einzelne Datei verarbeiten
	extractorObj := &extractor.ZUGFeRDExtractor{
		InputPath:  files[0],
		OutputPath: outputPath,
		Verbose:    verbose,
	}

	if err := extractorObj.ExtractXML(); err != nil {
		log.Fatalf("Fehler beim Extrahieren von XML: %v", err)
	}
}

func printUsage() {
	fmt.Println("ZUGFeRD XML Extractor v1.0")
	fmt.Println("Verwendung: zugferd-extractor [optionen] <pfad-zur-zugferd-pdf>")
	fmt.Println()
	fmt.Println("Optionen:")
	fmt.Println("  -v         Ausführliche Ausgabe")
	fmt.Println("  -o <pfad>  Ausgabepfad für die XML-Datei")
	fmt.Println("  -h         Diese Hilfe anzeigen")
	fmt.Println()
	fmt.Println("Beispiele:")
	fmt.Println("  zugferd-extractor rechnung.pdf")
	fmt.Println("  zugferd-extractor -v rechnung.pdf")
	fmt.Println("  zugferd-extractor -o ausgabe.xml rechnung.pdf")
	fmt.Println("  zugferd-extractor *.pdf")
	fmt.Println()
	fmt.Println("Unterstützte Formate:")
	fmt.Println("  - ZUGFeRD 1.0, 2.0, 2.1, 2.3")
	fmt.Println("  - Factur-X")
	fmt.Println("  - XRechnung")
}

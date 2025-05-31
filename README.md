# ZUGFeRD XML Extractor

Ein einfaches Command-Line-Tool zur Extraktion von ZUGFeRD/Factur-X/XRechnung XML-Daten aus PDF-Dateien.

## 📑 Übersicht

Dieses Tool extrahiert eingebettete XML-Dateien aus ZUGFeRD-konformen PDF-Dateien. Es unterstützt:

- ZUGFeRD 1.0, 2.0, 2.1, 2.3
- Factur-X
- XRechnung

## 🛠️ Installation

### Voraussetzungen

- Go 1.16 oder höher

### Kompilieren

```bash
# Repository klonen
git clone https://github.com/mmuyakwa/zugferd-extractor.git
cd zugferd-extractor

# Abhängigkeiten installieren 
go mod tidy

# Für aktuelles System bauen
go build -o zugferd-extractor ./cmd/zugferd-extractor

# Mit Build-Script für alle Plattformen
chmod +x build.sh
./build.sh
```

```bash
# Spezifische Pakete aktualisieren
go get github.com/pdfcpu/pdfcpu@v0.6.0
go get golang.org/x/image/ccitt
go get golang.org/x/image/webp
go get golang.org/x/text/unicode/norm

# Go-Module aktualisieren
go mod tidy
```

## 📋 Verwendung

### Einzelne Datei verarbeiten

```bash
./zugferd-extractor rechnung.pdf
```

### Mit ausführlicher Ausgabe

```bash
./zugferd-extractor rechnung.pdf -v
```

### Mit spezifischem Ausgabepfad

```bash
./zugferd-extractor -o ausgabe.xml rechnung.pdf
```

### Mehrere Dateien verarbeiten

```bash
./zugferd-extractor *.pdf
```

### Allgemeine Syntax

```bash
zugferd-extractor [optionen] <pfad-zur-zugferd-pdf>

Optionen:
  -v         Ausführliche Ausgabe
  -o <pfad>  Ausgabepfad für die XML-Datei
  -h         Diese Hilfe anzeigen
```

## 🔍 Funktionen

- Automatische Erkennung aller ZUGFeRD-XML-Varianten
- Parallel-Verarbeitung mehrerer Dateien
- Extraktion in sinnvoll benannte XML-Dateien
- Validierung des XML-Inhalts
- Detaillierter Verbose-Modus

## 🧰 Technologie

- 100% Open Source
- Basierend auf [pdfcpu](https://github.com/pdfcpu/pdfcpu)
- Keine externen Abhängigkeiten zur Laufzeit

## 📄 Lizenz

MIT

---

Entwickelt für die sichere und zuverlässige Extraktion elektronischer Rechnungsdaten.

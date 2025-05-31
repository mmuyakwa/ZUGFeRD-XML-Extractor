package validation

import (
    "strings"
)

// Validator is responsible for validating ZUGFeRD XML data
type Validator struct{}

// IsZUGFeRDXML checks if the XML data appears to be a ZUGFeRD document
func (v *Validator) IsZUGFeRDXML(data []byte) bool {
    if len(data) == 0 {
        return false
    }

    content := string(data)
    contentLower := strings.ToLower(content)

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

    for _, indicator := range indicators {
        if strings.Contains(contentLower, strings.ToLower(indicator)) {
            return true
        }
    }

    return false
}

// ValidateZUGFeRDXML performs additional validation on the XML content
func (v *Validator) ValidateZUGFeRDXML(data []byte) bool {
    content := string(data)

    hasXMLDecl := strings.Contains(content, "<?xml")
    hasRootElement := strings.Contains(content, "CrossIndustryDocument") || 
                     strings.Contains(content, "CrossIndustryInvoice")
    hasNamespace := strings.Contains(content, "xmlns:") && 
                   (strings.Contains(content, "urn:ferd:") || 
                    strings.Contains(content, "urn:cen.eu:en16931"))

    return hasXMLDecl && hasRootElement && hasNamespace
}
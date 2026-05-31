package utils

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// ExtractDOCXText reads a DOCX file and returns its text content
func ExtractDOCXText(filePath string) (string, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open DOCX: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open document.xml: %w", err)
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("failed to read document.xml: %w", err)
			}

			return extractTextFromXML(data)
		}
	}

	return "", fmt.Errorf("document.xml not found in DOCX")
}

func extractTextFromXML(data []byte) (string, error) {
	var sb strings.Builder
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	inText := false

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return sb.String(), nil // return what we have
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "t" {
				inText = true
			}
			if t.Name.Local == "p" {
				sb.WriteString("\n")
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inText = false
			}
		case xml.CharData:
			if inText {
				sb.Write(t)
			}
		}
	}

	return strings.TrimSpace(sb.String()), nil
}

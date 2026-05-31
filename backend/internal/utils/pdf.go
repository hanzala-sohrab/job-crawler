package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/dslipak/pdf"
)

// ExtractPDFText reads a PDF file and returns its text content
func ExtractPDFText(filePath string) (string, error) {
	r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}

	var sb strings.Builder
	totalPages := r.NumPage()

	for i := 1; i <= totalPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}

	result := sb.String()
	if strings.TrimSpace(result) == "" {
		// Fallback: try reading raw bytes
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("PDF text extraction failed and raw read failed: %w", err)
		}
		return string(data), nil
	}

	return result, nil
}

package converters

import (
	"testing"

	"github.com/shellcell/convert/internal/domain"
)

func TestCalibreDoesNotClaimPlainTextHTMLOrPDFOutput(t *testing.T) {
	converter := NewCalibre(nil)
	if converter.CanConvert(domain.FormatTXT, domain.FormatHTML) {
		t.Fatal("calibre should not claim txt -> html")
	}
	if converter.CanConvert(domain.FormatTXT, domain.FormatPDF) {
		t.Fatal("calibre should not claim txt -> pdf")
	}
	if !converter.CanConvert(domain.FormatTXT, domain.FormatEPUB) {
		t.Fatal("calibre should still support txt -> epub")
	}
}

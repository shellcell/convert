// Package scan holds small, dependency-free file inspection helpers shared
// by the filesystem adapter, the interactive prompt, and converter backends.
package scan

import (
	"encoding/xml"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

const textProbeSize = 8192

// LooksLikeText reports whether the byte sample reads as plain text.
func LooksLikeText(data []byte) bool {
	if len(data) == 0 {
		return true
	}

	control := 0
	for _, b := range data {
		if b == 0 {
			return false
		}
		if b < 0x20 && b != '\n' && b != '\r' && b != '\t' && b != '\f' && b != '\b' {
			control++
		}
		if b == 0x7f {
			control++
		}
	}
	return control*100 <= len(data)*30
}

// IsTextFile probes the beginning of the file and reports whether it looks
// like plain text. Directories are never text files.
func IsTextFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buffer := make([]byte, textProbeSize)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}
	return LooksLikeText(buffer[:n]), nil
}

// SVGSize returns the intrinsic size of an SVG document as "WxH", preferring
// width/height attributes and falling back to the viewBox.
func SVGSize(path string) (string, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", false, err
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return "", false, nil
		}
		if err != nil {
			return "", false, err
		}
		start, ok := token.(xml.StartElement)
		if !ok || !strings.EqualFold(start.Name.Local, "svg") {
			continue
		}

		attrs := map[string]string{}
		for _, attr := range start.Attr {
			attrs[strings.ToLower(attr.Name.Local)] = attr.Value
		}
		if width, ok := SVGLength(attrs["width"]); ok {
			if height, ok := SVGLength(attrs["height"]); ok {
				return SizeString(width, height), true, nil
			}
		}
		if width, height, ok := SVGViewBoxSize(attrs["viewbox"]); ok {
			return SizeString(width, height), true, nil
		}
		return "", false, nil
	}
}

// SVGLength parses an SVG length attribute value such as "128", "128px".
// Percent values are rejected because they have no absolute size.
func SVGLength(value string) (float64, bool) {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasSuffix(value, "%") {
		return 0, false
	}
	end := 0
	for end < len(value) {
		ch := value[end]
		if (ch >= '0' && ch <= '9') || ch == '.' || ch == '+' || ch == '-' {
			end++
			continue
		}
		break
	}
	if end == 0 {
		return 0, false
	}
	number, err := strconv.ParseFloat(value[:end], 64)
	if err != nil || number <= 0 {
		return 0, false
	}
	return number, true
}

// SVGViewBoxSize extracts width and height from a viewBox attribute value.
func SVGViewBoxSize(value string) (float64, float64, bool) {
	fields := strings.Fields(strings.ReplaceAll(strings.TrimSpace(value), ",", " "))
	if len(fields) != 4 {
		return 0, 0, false
	}
	width, err := strconv.ParseFloat(fields[2], 64)
	if err != nil || width <= 0 {
		return 0, 0, false
	}
	height, err := strconv.ParseFloat(fields[3], 64)
	if err != nil || height <= 0 {
		return 0, 0, false
	}
	return width, height, true
}

// SizeString renders float dimensions as a "WxH" string with a 1px floor.
func SizeString(width float64, height float64) string {
	return strconv.Itoa(max(1, int(math.Round(width)))) + "x" + strconv.Itoa(max(1, int(math.Round(height))))
}

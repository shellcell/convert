package converters

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shellcell/cnvrt/internal/domain"
)

func convertStructured(t *testing.T, name string, content string, input domain.Format, output domain.Format, options domain.ToolOptions) string {
	t.Helper()
	dir := t.TempDir()
	inputPath := filepath.Join(dir, name)
	outputPath := filepath.Join(dir, "out."+output.Extension())
	if err := os.WriteFile(inputPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	converter := NewStructuredData()
	_, err := converter.Convert(context.Background(), domain.ConvertJob{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		InputFormat:  input,
		OutputFormat: output,
		Options:      domain.ConvertOptions{ToolOptions: options},
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestStructuredCSVToMarkdownTable(t *testing.T) {
	got := convertStructured(t, "people.csv", "name,age\nAda,36\nGrace,85\n", domain.FormatCSV, domain.FormatMD, nil)
	want := "| age | name |\n| --- | --- |\n| 36 | Ada |\n| 85 | Grace |\n"
	if got != want {
		t.Fatalf("unexpected markdown:\n%s", got)
	}
}

func TestStructuredJSONToPlainText(t *testing.T) {
	got := convertStructured(t, "config.json", `{"name":"app","ports":[80,443]}`, domain.FormatJSON, domain.FormatTXT, nil)
	if !strings.Contains(got, "name: app") || !strings.Contains(got, "- 80") {
		t.Fatalf("unexpected plain text:\n%s", got)
	}
}

func TestStructuredJSONToRawText(t *testing.T) {
	content := `{"name":"app"}`
	options := domain.ToolOptions{"structured": {"text_style": []string{"raw"}}}
	got := convertStructured(t, "config.json", content, domain.FormatJSON, domain.FormatTXT, options)
	if strings.TrimSpace(got) != content {
		t.Fatalf("raw text should copy input, got:\n%s", got)
	}
}

func TestStructuredJSONLinesToJSON(t *testing.T) {
	got := convertStructured(t, "events.jsonl", "{\"id\":1}\n{\"id\":2}\n", domain.FormatJSONL, domain.FormatJSON, nil)
	if !strings.Contains(got, "\"id\": 1") || !strings.HasPrefix(strings.TrimSpace(got), "[") {
		t.Fatalf("unexpected json:\n%s", got)
	}
}

func TestStructuredEnvRoundTrip(t *testing.T) {
	got := convertStructured(t, "vars.env", "# comment\nexport HOST=localhost\nPORT=8080\nNAME=\"my app\"\n", domain.FormatENV, domain.FormatJSON, nil)
	for _, want := range []string{"\"HOST\": \"localhost\"", "\"PORT\": \"8080\"", "\"NAME\": \"my app\""} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %s in:\n%s", want, got)
		}
	}

	back := convertStructured(t, "vars.json", `{"HOST":"localhost","NAME":"my app"}`, domain.FormatJSON, domain.FormatENV, nil)
	if !strings.Contains(back, "HOST=localhost") || !strings.Contains(back, "NAME=\"my app\"") {
		t.Fatalf("unexpected env:\n%s", back)
	}
}

func TestStructuredTSVToCSV(t *testing.T) {
	got := convertStructured(t, "data.tsv", "a\tb\n1\t2\n", domain.FormatTSV, domain.FormatCSV, nil)
	if !strings.Contains(got, "a,b") || !strings.Contains(got, "1,2") {
		t.Fatalf("unexpected csv:\n%s", got)
	}
}

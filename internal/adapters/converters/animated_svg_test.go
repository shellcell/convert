package converters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shellcell/cnvrt/internal/domain"
)

func TestAnimatedSVGCapabilitiesOnlyForAnimatedSVG(t *testing.T) {
	dir := t.TempDir()
	animated := filepath.Join(dir, "animated.svg")
	static := filepath.Join(dir, "static.svg")
	if err := os.WriteFile(animated, []byte(`<svg width="10" height="10"><circle r="2"><animate attributeName="r" values="2;4" dur="1s"/></circle></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(static, []byte(`<svg width="10" height="10"><circle r="2"/></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}

	converter := NewAnimatedSVG(nil)
	if len(converter.CapabilitiesForInput(static, domain.FormatSVG)) != 0 {
		t.Fatal("static SVG should not expose animated outputs")
	}

	caps := converter.CapabilitiesForInput(animated, domain.FormatSVG)
	if !hasCapability(caps, domain.FormatSVG, domain.FormatMP4) || !hasCapability(caps, domain.FormatSVG, domain.FormatGIF) {
		t.Fatalf("animated SVG should expose video/animation outputs: %#v", caps)
	}
}

func TestSVGAnimationDurationUsesSMILTiming(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "duration.svg")
	if err := os.WriteFile(path, []byte(`<svg><circle><animate attributeName="r" begin="1s" dur="2s" repeatCount="3"/></circle></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}

	got := svgAnimationDuration(path, 3)
	if got != 7 {
		t.Fatalf("expected duration 7, got %v", got)
	}
}

func TestAnimatedSVGPreviewUsesEvenDimensionsForH264Outputs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "odd.svg")
	if err := os.WriteFile(path, []byte(`<svg width="1138" height="495"><circle><animate attributeName="r" dur="1s"/></circle></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}

	preview := NewAnimatedSVG(nil).PreviewCommands(domain.ConvertJob{
		InputPath:    path,
		OutputPath:   filepath.Join(dir, "odd.mp4"),
		InputFormat:  domain.FormatSVG,
		OutputFormat: domain.FormatMP4,
	})
	if !preview.Editable || preview.EditableCommand != 1 {
		t.Fatalf("expected editable ffmpeg command, got %#v", preview)
	}
	if len(preview.Commands) != 2 {
		t.Fatalf("expected browser and ffmpeg previews, got %#v", preview.Commands)
	}
	if !containsString(preview.Commands[0].Args, "--window-size=1138,496") {
		t.Fatalf("expected even window size, got %#v", preview.Commands[0].Args)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

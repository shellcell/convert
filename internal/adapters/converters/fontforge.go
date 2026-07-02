package converters

import (
	"context"

	"github.com/shellcell/convert/internal/domain"
	"github.com/shellcell/convert/internal/ports"
)

type FontForge struct {
	runner ports.CommandRunner
	caps   []domain.ConversionCapability
}

func NewFontForge(runner ports.CommandRunner) *FontForge {
	formats := []domain.Format{
		domain.FormatTTF,
		domain.FormatOTF,
		domain.FormatWOFF,
		domain.FormatWOFF2,
		domain.FormatEOT,
		domain.FormatBDF,
		domain.FormatPCF,
		domain.FormatFON,
		domain.FormatPFA,
		domain.FormatPFB,
	}
	return &FontForge{runner: runner, caps: capabilities(formats, formats, 80, false, false)}
}

func (c *FontForge) ID() string { return "fontforge" }

func (c *FontForge) RequiredCommands() []string { return []string{"fontforge"} }

func (c *FontForge) Capabilities() []domain.ConversionCapability {
	return append([]domain.ConversionCapability(nil), c.caps...)
}

func (c *FontForge) CanConvert(input domain.Format, output domain.Format) bool {
	return hasCapability(c.caps, input, output)
}

func (c *FontForge) Convert(ctx context.Context, job domain.ConvertJob) (domain.ConversionResult, error) {
	return runSimple(ctx, c.runner, "fontforge", c.args(job), job, c.ID())
}

func (c *FontForge) PreviewCommands(job domain.ConvertJob) ports.CommandPreview {
	return previewCommand("fontforge", c.args(job))
}

func (c *FontForge) args(job domain.ConvertJob) []string {
	script := `Open($1); Generate($2); Close();`
	return []string{"-lang=ff", "-c", script, job.InputPath, job.OutputPath}
}

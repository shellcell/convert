package converters

import (
	"context"

	"github.com/shellcell/convert/internal/domain"
	"github.com/shellcell/convert/internal/ports"
)

type FFmpeg struct {
	runner ports.CommandRunner
	caps   []domain.ConversionCapability
}

func NewFFmpeg(runner ports.CommandRunner) *FFmpeg {
	inputs := []domain.Format{
		domain.FormatMP4,
		domain.FormatMOV,
		domain.FormatAVI,
		domain.FormatWebM,
		domain.FormatMKV,
		domain.FormatM4V,
		domain.FormatMPG,
		domain.FormatMPEG,
		domain.FormatFLV,
		domain.FormatOGV,
		domain.FormatGIF,
		domain.FormatAPNG,
		domain.FormatWebP,
		domain.FormatMP3,
		domain.FormatWAV,
		domain.FormatFLAC,
		domain.FormatAAC,
		domain.FormatM4A,
		domain.FormatOGG,
		domain.FormatOPUS,
		domain.FormatWMA,
		domain.FormatAIFF,
	}
	outputs := []domain.Format{
		domain.FormatMP4,
		domain.FormatWebM,
		domain.FormatGIF,
		domain.FormatWebP,
		domain.FormatMOV,
		domain.FormatAVI,
		domain.FormatMKV,
		domain.FormatM4V,
		domain.FormatMP3,
		domain.FormatWAV,
		domain.FormatFLAC,
		domain.FormatAAC,
		domain.FormatM4A,
		domain.FormatOGG,
		domain.FormatOPUS,
		domain.FormatAIFF,
	}
	return &FFmpeg{runner: runner, caps: capabilities(inputs, outputs, 90, true, true)}
}

func (c *FFmpeg) ID() string { return "ffmpeg" }

func (c *FFmpeg) RequiredCommands() []string { return []string{"ffmpeg"} }

func (c *FFmpeg) Capabilities() []domain.ConversionCapability {
	return append([]domain.ConversionCapability(nil), c.caps...)
}

func (c *FFmpeg) CanConvert(input domain.Format, output domain.Format) bool {
	return hasCapability(c.caps, input, output)
}

func (c *FFmpeg) Convert(ctx context.Context, job domain.ConvertJob) (domain.ConversionResult, error) {
	return runSimple(ctx, c.runner, "ffmpeg", c.args(job), job, c.ID())
}

func (c *FFmpeg) PreviewCommands(job domain.ConvertJob) ports.CommandPreview {
	return previewCommand("ffmpeg", c.args(job))
}

func (c *FFmpeg) args(job domain.ConvertJob) []string {
	args := []string{"-hide_banner", "-loglevel", "error"}
	if job.Options.Overwrite {
		args = append(args, "-y")
	} else {
		args = append(args, "-n")
	}
	args = append(args, "-i", job.InputPath)
	if ffmpegNeedsEvenVideo(job.InputFormat, job.OutputFormat) {
		args = append(args, "-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2", "-pix_fmt", "yuv420p")
	}
	args = append(args, job.OutputPath)
	return args
}

func ffmpegNeedsEvenVideo(input domain.Format, output domain.Format) bool {
	if input.IsAudio() {
		return false
	}
	switch output {
	case domain.FormatMP4, domain.FormatMOV, domain.FormatMKV, domain.FormatM4V:
		return true
	default:
		return false
	}
}

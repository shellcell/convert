package converters

import (
	"slices"
	"strings"
	"testing"

	"github.com/shellcell/cnvrt/internal/domain"
)

func TestFFmpegDoesNotDeclareAudioToVideo(t *testing.T) {
	converter := NewFFmpeg(nil)
	if converter.CanConvert(domain.FormatMP3, domain.FormatMP4) {
		t.Fatal("mp3 -> mp4 should not be declared")
	}
	if converter.CanConvert(domain.FormatWAV, domain.FormatGIF) {
		t.Fatal("wav -> gif should not be declared")
	}
	if !converter.CanConvert(domain.FormatMP4, domain.FormatMP3) {
		t.Fatal("mp4 -> mp3 audio extraction should be declared")
	}
	if !converter.CanConvert(domain.FormatGIF, domain.FormatMP4) {
		t.Fatal("gif -> mp4 should be declared")
	}
}

func TestFFmpegAudioExtractionArgs(t *testing.T) {
	converter := NewFFmpeg(nil)
	args := converter.args(domain.ConvertJob{
		InputPath:    "in.mp4",
		OutputPath:   "out.mp3",
		InputFormat:  domain.FormatMP4,
		OutputFormat: domain.FormatMP3,
	})
	if !slices.Contains(args, "-vn") {
		t.Fatalf("expected -vn for audio extraction, got %v", args)
	}
	index := slices.Index(args, "-b:a")
	if index < 0 || args[index+1] != "192k" {
		t.Fatalf("expected default audio bitrate, got %v", args)
	}
}

func TestFFmpegGIFUsesPalette(t *testing.T) {
	converter := NewFFmpeg(nil)
	args := converter.args(domain.ConvertJob{
		InputPath:    "in.mp4",
		OutputPath:   "out.gif",
		InputFormat:  domain.FormatMP4,
		OutputFormat: domain.FormatGIF,
	})
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "palettegen") || !strings.Contains(joined, "fps=15") {
		t.Fatalf("expected palette filter with default fps, got %v", args)
	}
}

func TestFFmpegResizeAndCompress(t *testing.T) {
	converter := NewFFmpeg(nil)
	args := converter.args(domain.ConvertJob{
		InputPath:    "in.mov",
		OutputPath:   "out.mp4",
		InputFormat:  domain.FormatMOV,
		OutputFormat: domain.FormatMP4,
		Options: domain.ConvertOptions{
			Action:  domain.ActionCompress,
			Quality: 85,
			Resize:  "800x",
		},
	})
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "scale=800:-2") {
		t.Fatalf("expected even-safe scale filter, got %v", args)
	}
	index := slices.Index(args, "-crf")
	if index < 0 || args[index+1] != "23" {
		t.Fatalf("expected crf 23 for quality 85, got %v", args)
	}
}

func TestFFmpegExtraArgsPassThrough(t *testing.T) {
	converter := NewFFmpeg(nil)
	args := converter.args(domain.ConvertJob{
		InputPath:    "in.wav",
		OutputPath:   "out.flac",
		InputFormat:  domain.FormatWAV,
		OutputFormat: domain.FormatFLAC,
		Options: domain.ConvertOptions{
			ToolOptions: domain.ToolOptions{"ffmpeg": {"args": []string{"-compression_level 8"}}},
		},
	})
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-compression_level 8") {
		t.Fatalf("expected pass-through args, got %v", args)
	}
}

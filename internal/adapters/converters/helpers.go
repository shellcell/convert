package converters

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/shellcell/cnvrt/internal/domain"
	"github.com/shellcell/cnvrt/internal/ports"
	"github.com/shellcell/cnvrt/internal/shell"
)

func hasCapability(capabilities []domain.ConversionCapability, input domain.Format, output domain.Format) bool {
	for _, capability := range capabilities {
		if capability.Input == input && capability.Output == output {
			return true
		}
	}
	return false
}

func capabilities(inputs []domain.Format, outputs []domain.Format, priority int, lossy bool, preservesAnimation bool) []domain.ConversionCapability {
	result := make([]domain.ConversionCapability, 0, len(inputs)*len(outputs))
	for _, input := range inputs {
		for _, output := range outputs {
			result = append(result, domain.ConversionCapability{
				Input:              input,
				Output:             output,
				Priority:           priority,
				Lossy:              lossy,
				PreservesAnimation: preservesAnimation,
			})
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Input == result[j].Input {
			return result[i].Output < result[j].Output
		}
		return result[i].Input < result[j].Input
	})
	return result
}

func qualityArgs(job domain.ConvertJob) []string {
	if job.Options.Quality <= 0 {
		return nil
	}
	return []string{"-quality", strconv.Itoa(job.Options.Quality)}
}

func resizeDimensions(value string) (string, string) {
	value = strings.TrimSpace(value)
	parts := strings.SplitN(value, "x", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

// stringOption returns the first configured value for (tool, key), or the
// fallback when unset.
func stringOption(options domain.ToolOptions, tool string, key string, fallback string) string {
	values := options.Values(tool, key)
	if len(values) == 0 || strings.TrimSpace(values[0]) == "" {
		return fallback
	}
	return strings.TrimSpace(values[0])
}

func intOption(options domain.ToolOptions, tool string, key string, fallback int) int {
	value, err := strconv.Atoi(stringOption(options, tool, key, ""))
	if err != nil {
		return fallback
	}
	return value
}

func floatOption(options domain.ToolOptions, tool string, key string, fallback float64) float64 {
	value, err := strconv.ParseFloat(stringOption(options, tool, key, ""), 64)
	if err != nil {
		return fallback
	}
	return value
}

// extraArgs returns user-supplied pass-through arguments for a backend: every
// value under `<tool>.args` is split on whitespace and appended to the
// command line, so any tool flag stays reachable without code changes.
func extraArgs(options domain.ToolOptions, tool string) []string {
	var args []string
	for _, value := range options.Values(tool, "args") {
		args = append(args, strings.Fields(value)...)
	}
	return args
}

func commandError(command ports.Command, result ports.CommandResult, err error) error {
	return shell.CommandError(command, result, err)
}

func commandStringError(command string, result ports.CommandResult, err error) error {
	return shell.CommandStringError(command, result, err)
}

func shellCommand(command string) ports.Command {
	return shell.Command(command)
}

func commandLine(command ports.Command) string {
	return shell.Line(command)
}

func runSimple(ctx context.Context, runner ports.CommandRunner, command string, args []string, job domain.ConvertJob, backend string) (domain.ConversionResult, error) {
	cmd := ports.Command{Name: command, Args: args}
	result, err := runner.Run(ctx, cmd)
	if err != nil {
		return domain.ConversionResult{}, commandError(cmd, result, err)
	}

	return domain.ConversionResult{Job: job, Backend: backend, OutputPath: job.OutputPath}, nil
}

func previewCommand(command string, args []string) ports.CommandPreview {
	return ports.CommandPreview{Commands: []ports.Command{{Name: command, Args: args}}, Editable: true}
}

func moveFile(from string, to string) error {
	if err := os.Rename(from, to); err == nil {
		return nil
	}

	input, err := os.Open(from)
	if err != nil {
		return err
	}
	defer input.Close()

	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}

	output, err := os.Create(to)
	if err != nil {
		return err
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return err
	}

	return os.Remove(from)
}

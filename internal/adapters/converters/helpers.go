package converters

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/shellcell/convert/internal/domain"
	"github.com/shellcell/convert/internal/ports"
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

func commandError(command ports.Command, result ports.CommandResult, err error) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf("command: %s", commandLine(command))
	if result.Stderr != "" {
		return fmt.Errorf("%s: %w: %s", message, err, result.Stderr)
	}
	if result.Stdout != "" {
		return fmt.Errorf("%s: %w: %s", message, err, result.Stdout)
	}
	return fmt.Errorf("%s: %w", message, err)
}

func commandStringError(command string, result ports.CommandResult, err error) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf("command: %s", command)
	if result.Stderr != "" {
		return fmt.Errorf("%s: %w: %s", message, err, result.Stderr)
	}
	if result.Stdout != "" {
		return fmt.Errorf("%s: %w: %s", message, err, result.Stdout)
	}
	return fmt.Errorf("%s: %w", message, err)
}

func shellCommand(command string) ports.Command {
	if runtime.GOOS == "windows" {
		return ports.Command{Name: "cmd", Args: []string{"/C", command}}
	}
	return ports.Command{Name: "sh", Args: []string{"-c", command}}
}

func commandLine(command ports.Command) string {
	parts := []string{command.Name}
	parts = append(parts, command.Args...)
	for i, part := range parts {
		if part == "" || strings.ContainsAny(part, " \t\n\"'\\$`") {
			parts[i] = strconv.Quote(part)
		}
	}
	return strings.Join(parts, " ")
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

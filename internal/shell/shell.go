// Package shell holds shared helpers for rendering, quoting, and wrapping
// external commands. It is used by the application core, converter adapters,
// and config-defined tools so command handling stays consistent.
package shell

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/shellcell/convert/internal/ports"
)

// Command wraps a raw command line for execution through the platform shell.
func Command(command string) ports.Command {
	if runtime.GOOS == "windows" {
		return ports.Command{Name: "cmd", Args: []string{"/C", command}}
	}
	return ports.Command{Name: "sh", Args: []string{"-c", command}}
}

// Quote quotes a single argument for POSIX shells. Unlike strconv.Quote it
// uses single quotes, so `$`, backticks, and backslashes stay literal.
func Quote(value string) string {
	if value != "" && !strings.ContainsAny(value, " \t\n\"'\\$`;&|<>!*?[]{}()~#") {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

// Line renders a command as a single shell-quoted line.
func Line(command ports.Command) string {
	var b strings.Builder
	if command.Dir != "" {
		b.WriteString("cd ")
		b.WriteString(Quote(command.Dir))
		b.WriteString(" && ")
	}
	b.WriteString(Quote(command.Name))
	for _, arg := range command.Args {
		b.WriteByte(' ')
		b.WriteString(Quote(arg))
	}
	return b.String()
}

// CommandError wraps a failed command execution with the rendered command
// line and any captured output.
func CommandError(command ports.Command, result ports.CommandResult, err error) error {
	return commandError(Line(command), result, err)
}

// CommandStringError is CommandError for commands given as raw strings.
func CommandStringError(command string, result ports.CommandResult, err error) error {
	return commandError(command, result, err)
}

func commandError(command string, result ports.CommandResult, err error) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf("command: %s", command)
	if result.Stderr != "" {
		return fmt.Errorf("%s: %w: %s", message, err, strings.TrimSpace(result.Stderr))
	}
	if result.Stdout != "" {
		return fmt.Errorf("%s: %w: %s", message, err, strings.TrimSpace(result.Stdout))
	}
	return fmt.Errorf("%s: %w", message, err)
}

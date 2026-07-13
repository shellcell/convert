# Architecture

`cnvrt` uses clean layered architecture. The CLI and external tools are adapters. The application core owns conversion decisions, batch reporting, dependency checks, and output naming rules.

## Dependency Direction

```text
cmd/cnvrt
  -> internal/bootstrap
  -> internal/adapters
  -> internal/app
  -> internal/ports
  -> internal/domain
```

Domain imports nothing from app, adapters, CLI, terminal libraries, filesystem code, or external command code.

## Packages

```text
cmd/cnvrt/main.go
internal/domain              conversion concepts and the format catalog
internal/ports               interfaces owned by the core
internal/app                 use cases, policies, reports
internal/bootstrap           wiring: builds adapters and the service
internal/adapters/cli        flag/command parsing, final report rendering
internal/adapters/prompt     interactive selectors and confirmations
internal/adapters/fs         directory listing and path checks
internal/adapters/exec       external command execution
internal/adapters/install    install hints for missing tools
internal/adapters/settings   user settings and backend preferences
internal/adapters/toolconfig config-defined formats and tools
internal/adapters/progress   per-job progress output
internal/adapters/converters built-in backends
internal/scan                dependency-free file inspection helpers
internal/shell               shared command rendering/quoting helpers
internal/theme               color palettes
```

`internal/scan`, `internal/shell`, and `internal/theme` are leaf helper packages that any layer may import; they depend on nothing inside the project.

## Domain Layer

The domain layer defines stable conversion concepts:

- `Format`
- `FileRef`
- `ConvertJob`
- `ConvertOptions`
- `ToolOptions`
- `ConversionCapability`
- `ConversionResult`
- `TransformAction`
- `ArchiveAction`
- `MissingDependencyError` and other domain errors

Formats are open-ended. Built-ins are registered in code, but unknown normalized format names are accepted so config-defined tools can introduce new formats without recompilation.

## Application Layer

The app layer owns use cases and policies:

- interactive conversion flow after prompt choices are returned
- direct conversion flow
- tolerant batch conversion
- job building and validation
- output name generation
- same-format action handling
- archive extraction/compression routing
- backend selection and user backend preferences
- command preview confirmation before a job starts
- dependency status, runtime availability, and install hints
- output format availability choices for interactive prompts
- final run report data

Important types:

- `Service`
- `ConvertRequest`
- `InteractiveRequest`
- `Preferences` and `PairPreference`
- `RunReport`
- `JobReport`
- `DependencyStatus`
- `FormatChoice`

Batch conversion does not abort on the first unsupported pair. Each input becomes a report item with status `converted`, `skipped`, or `failed`.

## Ports

Ports are interfaces owned by the core. Adapters implement them.

Core ports:

- `Converter`
- `FileDiscovery`
- `FileSystem`
- `Prompt`
- `CommandRunner`
- `InstallAdvisor`
- `ProgressReporter`
- `AppRunner`

Optional interfaces a converter may also implement:

- `RuntimeDependencyAware` — declares sub-dependencies beyond `PATH` lookups, such as Pandoc PDF engines
- `InputCapabilityAware` — derives capabilities from a specific input file
- `DependencyStatusAware` — reports custom `doctor` rows
- `CommandPreviewer` — exposes the command that will run, for preview/confirmation
- `CommandOverrideConverter` — runs a user-edited command inside the normal pipeline
- `OptionsAware` — declares interactive/`--opt` conversion settings
- `Describable` — supplies the description shown by `backends`

## Adapters

### CLI

The CLI adapter parses flags and commands, calls the application service, and renders final status reports.

Commands:

```text
doctor      backend availability, purpose, and install hints
formats     known formats grouped by category
backends    each backend's description, commands, and pairs
config      interactively edit and save user settings
add-format  add a config-defined format
add-tool    add a config-defined converter tool
help        usage
```

Flags:

```text
-i, --input-format   override input format
-o, --output-format  output format
    --out-dir        output directory
    --overwrite      overwrite existing output files
    --quality        best-effort quality value for supported backends
    --compress       compress same-format image output
    --resize         resize value for supported image and video backends
    --action         same-format action: convert, compress, or resize
    --opt            backend option as tool.key=value, repeatable
```

Supported command shapes:

```bash
cnvrt
cnvrt input.svg output.png
cnvrt input.svg
cnvrt -i svg -o png abc.svg cde.svg
cnvrt -o png
cnvrt -o png ../../
cnvrt -o zip ./directory
cnvrt archive.tar.gz
```

### Prompt

The prompt adapter is built on Bubble Tea v2 models, styled with Lip Gloss v2. It renders inline, not as a full-screen TUI, and keeps a numeric stdin/stdout fallback for non-terminal input.

Terminal controls:

- up/down or j/k moves through options
- pageup/ctrl+up/option+up and pagedown/ctrl+down/option+down move by one visible page
- gg/ctrl+a jumps to the beginning
- G/shift+g/ctrl+e jumps to the end
- space or x selects/toggles where selection is available
- enter/right opens directories in the file picker, left goes up one level, and enter confirms other selections
- `/` filters options; esc clears the active filter
- a selects all currently filtered file-picker entries; c clears the selection
- s cycles list sorting (name, reverse, size, format)
- mouse wheel scrolls
- q aborts the interactive prompt

The file picker shows supported files and directories. Directories can be opened with enter and selected with space/x, which supports directory-to-archive workflows. It preserves cursor, scroll offset, and filter text per visited directory, so going deeper and back restores the previous position.

The output format picker receives `FormatChoice` values from the application layer. Unavailable formats are still displayed and dimmed with the missing dependency reason, but remain selectable. If selected, conversion continues into normal backend selection so the final report can include install hints for missing tools.

Before each terminal job starts, the prompt adapter shows the selected backend and command preview. Option flags are split onto indented colored lines to avoid unreadable wrapping. Supported previews can choose which command is editable; edited commands run through the shell for that job. A batch of multiple files is confirmed once. Non-terminal runs proceed without the confirmation prompt.

If a generated output path already exists and overwrite is not enabled, terminal mode asks for a different output path. Non-terminal mode keeps the strict failure behavior.

Interactive flow for `cnvrt`:

```text
list files from current directory
-> select one or more inputs
-> select output format
-> select whether outputs go to current directory or source directories
-> optionally select same-format action
-> optionally set backend options
-> confirm command preview
-> convert supported jobs
-> print report
```

Interactive flow for a single archive:

```text
detect archive
-> offer extract, choose output format, or cancel
```

Interactive same-format image flow:

```text
input format equals output format
-> offer compress, resize, or convert/copy
```

### Filesystem

The filesystem adapter lists directory entries, checks paths, detects directories, and creates output directories.

### Exec

The exec adapter wraps `os/exec`, captures stdout/stderr, normalizes exit failures, and caches `PATH` lookups.

### Progress

The progress adapter prints per-job start/success/skip/failure lines while a batch runs. The core uses a no-op reporter when none is supplied.

### Install Advisor

The install adapter suggests installation commands for missing converter tools.

Preference order:

- Use Homebrew if `brew` exists, even on Linux.
- Otherwise use the native package manager detected on the host.
- Fall back to configured ecosystem install hints or a manual hint.

Supported config managers:

```text
brew, apt, dnf, yum, pacman, zypper, apk,
npm, pnpm, yarn, pipx, pip, go, cargo, gem, composer
```

### Settings

The settings adapter loads user preferences and writes the file used by `cnvrt config`.

Load paths, later entries overriding earlier ones:

```text
~/.config/cnvrt/settings.json
./cnvrt.settings.json
CNVRT_SETTINGS paths
```

`cnvrt config` writes to `~/.config/cnvrt/settings.json`, merging into existing keys.

Schema:

```json
{
  "theme": "nord",
  "ui": { "show_help": true },
  "colors": {},
  "category_colors": {},
  "backends": { "svg->png": "resvg", "pdf": "ghostscript" },
  "tools": { "ffmpeg": { "audio_bitrate": "320k" } },
  "pairs": [
    { "input": "svg", "output": "png", "tools": ["resvg"] }
  ]
}
```

`backends` pins a preferred backend per `input->output` pair or per input format, which skips the interactive backend question. `tools` and `pairs` set default backend options globally or per pair. Themes are `nord` (default) and `classic`.

### Tool Config

The tool config adapter loads and writes JSON config files for additional formats and tools.

Load paths:

```text
./cnvrt.tools.json
./cnvrt.tools.d/*.json
~/.config/cnvrt/tools.json
~/.config/cnvrt/tools.d/*.json
CNVRT_TOOL_CONFIG paths
```

`cnvrt add-format` and `cnvrt add-tool` write generated files to `~/.config/cnvrt/tools.d/*.json`.

## Built-In Converter Adapters

In-process, no external command required:

- `structured`: JSON, JSONL, YAML, TOML, CSV, TSV, INI, ENV, XML, PLIST conversions, plus plain-text/Markdown rendering
- `archive`: stdlib zip/tar/tar.gz/tar.bz2 extraction and zip/tar/tar.gz creation

External-tool backends:

- `resvg`: SVG rendering
- `animated-svg`: animated SVG rendering through a headless browser and FFmpeg
- `imagemagick`: broad image conversion, resize, and quality options
- `ffmpeg`: video, audio, GIF, animated images
- `libreoffice`: office documents
- `pandoc`: text, Markdown, HTML, RTF, TeX, DOCX, EPUB
- `calibre`: EPUB, MOBI, AZW3, FB2, e-book workflows
- `gdal`: geo/vector map conversions
- `qemu-img`: VM and disk images
- `7z`: broad archive/package extraction and 7z/zip creation
- `fontforge`: outline and bitmap font conversions
- `inkscape`: SVG to pdf/eps/ps
- `ghostscript`: PostScript/PDF conversion and PDF re-compression
- `djvulibre`: DjVu to pdf/tiff/txt
- `poppler`: PDF text layer extraction
- `tesseract`: OCR to txt or searchable PDF
- `graphviz`: DOT diagrams
- `mermaid`: Mermaid diagrams
- `jupyter`: notebooks to html/markdown/script/pdf
- `tts`: text to speech
- `whisper`: speech to text

Config-defined converters are appended after built-ins.

## Backend Selection

Converters are ranked by declared capability priority for the pair, highest first, keeping registration order between equal priorities. User `backends` preferences reorder that list. The app then selects the first backend that:

- declares a matching `input -> output` capability
- has all required commands available on `PATH`
- has runtime sub-dependencies available when the backend declares them

When several installed backends can handle the pair, interactive mode asks once per pair per run. Non-terminal runs take the top-ranked backend.

If no backend supports a pair, the job is `skipped`. If a backend supports the pair but dependencies are missing, the job is `failed` with install hints.

## Input Detection

Input format detection prefers explicit `-i/--input-format`, then directories, then registered file extensions. If a file has no extension or only an unknown unsupported extension and its content looks like text, it is treated as `txt`. Custom formats with declared converter capabilities are preserved.

## Converter Options

Interactive converter options should default to not changing source parameters when the option is optional. SVG raster/video output size follows this rule: the user is first asked whether to set a size. If skipped, converters preserve source dimensions when possible. If enabled, the input defaults to source dimensions from SVG `width`/`height`, SVG `viewBox`, or readable raster image metadata, falling back to `1024x1024`.

The same options are scriptable via `--opt tool.key=value` and can be defaulted in settings.

## Output Naming

For direct conversion, the explicit output path is used:

```bash
cnvrt input.svg output.png
```

For format-based conversion, the default output is written to the current directory:

```bash
cnvrt -o png abc.svg   # -> abc.png
```

Interactive mode asks whether generated outputs should be saved in the current directory or beside each source. Direct non-interactive format-based conversion uses the current directory unless `--out-dir` is set.

Same-format image actions, archive extraction, and directory archive creation:

```text
photo.jpg        -> photo.compressed.jpg
photo.jpg        -> photo.resized.jpg
archive.tar.gz   -> archive/
./dist (-o zip)  -> dist.zip
```

Existing outputs are rejected unless `--overwrite` is set; in terminal mode the prompt asks for a different path instead.

## Config Schema

Example tool config:

```json
{
  "formats": [
    {
      "name": "plantuml",
      "aliases": ["puml"],
      "extensions": ["puml"],
      "category": "diagram"
    }
  ],
  "tools": [
    {
      "id": "plantuml",
      "commands": ["plantuml"],
      "capabilities": [
        { "input": "plantuml", "output": "svg", "priority": 80 }
      ],
      "convert": {
        "command": "plantuml",
        "args": ["-tsvg", "{input}"]
      },
      "install": [
        { "manager": "brew", "package": "plantuml" },
        { "manager": "apt", "package": "plantuml" },
        { "manager": "npm", "package": "node-plantuml" }
      ]
    }
  ]
}
```

Command placeholders:

```text
{input}
{output}
{input_format}
{output_format}
{quality}
{resize}
{action}
```

## Format Strategy

The built-in format catalog is intentionally broad because this tool targets developers and automation-heavy workflows.

Categories include images and vectors, documents and office files, e-books and scanned documents, audio and video, archives and package archives, fonts including bitmap fonts, geo/map formats, serialization/data formats, API/schema formats, and VM/disk images.

The important rule is that format recognition and conversion capability are separate. A format can be known without having a currently installed backend.

## Roadmap

Completed:

- clean architecture skeleton under `github.com/shellcell/cnvrt`
- interactive CLI with Bubble Tea v2 selectors and Lip Gloss v2 reports
- direct conversion, tolerant batch conversion, final status reports
- archive extraction and directory archive creation
- same-format image action prompts
- backend selection with user preferences and command preview/edit
- dependency install hints and `doctor`
- config-defined formats and tools, with interactive generators
- user settings, themes, and scriptable backend options

Next:

- dedicated compressor port and compressor adapters
- recursive discovery
- dry-run conversion plans
- integration tests for installed tools
- config validation command
- richer shell-template support for config-defined tools
- browser backend for HTML screenshots/PDF
- OCR for scanned PDFs (rasterize + tesseract, or ocrmypdf)
- mouse click selection in pickers (needs reliable view-relative mouse
  coordinates in inline mode; wheel scrolling already works)
- piper/whisper.cpp engines for TTS/STT alongside say/espeak-ng and
  openai-whisper

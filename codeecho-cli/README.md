# 📢 CodeEcho
*"Let your code speak back."*

CodeEcho is an open-source CLI tool that scans your repository and packages it into a single AI-friendly file. Perfect for feeding into ChatGPT, Claude, or any LLM.

Right now, CodeEcho supports **scanning repos** and outputting the result in Markdown, JSON, or plain text.

---

## Features (Stage 1)
- Scan an entire repo or folder.
- Skip common directories (`.git`, `node_modules`, `vendor`).
- Output in multiple formats:
  - Markdown (great for humans)
  - JSON (great for AIs/tools)
  - Text (minimalist raw mode)
- Flags for output file, format, and excluded dirs.

---

## Features

- **Repository Scanning**: Extract file structure and content from any directory
- **Multiple Output Formats**: XML, JSON, and Markdown support
- **File Processing**: Remove comments, compress code, strip empty lines
- **Smart Filtering**: Include/exclude files and directories based on patterns
- **Documentation Generation**: Auto-generate README, API docs, and project overviews
- **Cross-Platform**: Works on Linux, macOS, and Windows

## Installation

### Download Binary
Download the latest release for your platform from the [releases page](https://github.com/opskraken/code-echo/releases).

### Build from Source
```bash
git clone https://github.com/opskraken/code-echo.git
cd code-echo/codeecho-cli
go build -o codeecho main.go
```

### Install to System PATH (Optional)
```bash
# Linux/macOS
sudo mv codeecho /usr/local/bin/

# Windows
# Move codeecho.exe to a directory in your PATH
```

## Quick Start

```bash
# Basic repository scan (generates XML file)
codeecho scan .

# Scan with comment removal and code compression
codeecho scan . --remove-comments --compress-code

# Generate JSON output
codeecho scan . --format json

# Generate project documentation
codeecho doc .

# Show version information
codeecho version
```

## Commands

### `scan` - Repository Scanning

Scan a repository and generate AI-ready context files.

```bash
codeecho scan [path] [flags]
```

**Examples:**
```bash
codeecho scan .                                    # Scan current directory
codeecho scan /path/to/project                     # Scan specific directory
codeecho scan . -o my-repo.xml                     # Save to specific file
codeecho scan . --format json --no-content         # Structure-only JSON
```

#### Output Format Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--format, -f` | string | `xml` | Output format: xml, json, markdown |
| `--output, -o` | string | auto-generated | Output file path |
| `--include-summary` | bool | `true` | Include explanatory file summary |
| `--include-tree` | bool | `true` | Include directory structure |
| `--line-numbers` | bool | `false` | Show line numbers in code blocks |
| `--parsable` | bool | `true` | Use structured format tags |

#### File Processing Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--compress-code` | bool | `false` | Remove unnecessary whitespace |
| `--remove-comments` | bool | `false` | Strip comments from source files |
| `--remove-empty-lines` | bool | `false` | Remove blank lines |

#### File Filtering Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--content` | bool | `true` | Include file contents (use `--no-content` for structure only) |
| `--exclude-dirs` | strings | See below | Directories to exclude |
| `--include-exts` | strings | See below | File extensions to include |

**Default Excluded Directories:**
`.git`, `node_modules`, `vendor`, `.vscode`, `.idea`, `target`, `build`, `dist`

**Default Included Extensions:**
`.go`, `.js`, `.ts`, `.jsx`, `.tsx`, `.json`, `.md`, `.html`, `.css`, `.py`, `.java`, `.cpp`, `.c`, `.h`, `.rs`, `.rb`, `.php`, `.yml`, `.yaml`, `.toml`, `.xml`

#### Advanced Examples

```bash
# Remove comments and compress code
codeecho scan . --remove-comments --compress-code

# Structure-only scan (no file contents)
codeecho scan . --no-content

# Scan only Go files
codeecho scan . --include-exts .go

# Exclude additional directories
codeecho scan . --exclude-dirs .git,node_modules,target,tmp

# All processing options with custom output
codeecho scan . --remove-comments --remove-empty-lines --compress-code -o clean-repo.xml

# Minimal output (no summary or directory tree)
codeecho scan . --no-include-summary --no-include-tree
```

### `doc` - Documentation Generation

Generate documentation from repository analysis.

```bash
codeecho doc [path] [flags]
```

#### Documentation Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output, -o` | string | auto-generated | Output file path |
| `--type, -t` | string | `readme` | Documentation type: readme, api, overview |

**Examples:**
```bash
codeecho doc .                          # Generate README.md
codeecho doc . --type api               # Generate API.md
codeecho doc . --type overview          # Generate OVERVIEW.md
codeecho doc . -o CUSTOM-DOCS.md        # Custom output file
```

### `version` - Version Information

Display version and build information.

```bash
codeecho version
```

## Output Files

### Auto-Generated Filenames

When no `--output` flag is specified, files are automatically named using the pattern:
```
{project-name}-{processing-options}-{timestamp}.{extension}
```

**Examples:**
- `my-project-20250128-143022.xml` - Basic scan
- `my-project-no-comments-compressed-20250128-143025.xml` - Processed scan
- `my-project-structure-only-20250128-143028.xml` - Structure-only scan
- `my-project-20250128-143030.json` - JSON format

### Output Formats

#### XML Format (Default)
Structured XML similar to Repomix format, optimized for AI consumption.

#### JSON Format
Machine-readable JSON with complete file metadata and content.

#### Markdown Format
Human-readable documentation with syntax highlighting.

## Use Cases

### AI Context Generation
```bash
# Generate comprehensive context for AI tools
codeecho scan . --remove-comments --compress-code -o project-context.xml
```

### Code Review Preparation
```bash
# Clean, compressed version for review
codeecho scan . --remove-comments --remove-empty-lines --compress-code
```

### Project Analysis
```bash
# Structure and statistics only
codeecho scan . --no-content --format json -o project-analysis.json
```

### Documentation Generation
```bash
# Auto-generate project README
codeecho doc . --type readme
```

## Configuration

### Custom File Extensions
```bash
# Include specific file types
codeecho scan . --include-exts .go,.py,.js,.md

# Include all files (remove default filtering)
codeecho scan . --include-exts ""
```

### Custom Directory Exclusions
```bash
# Exclude additional directories
codeecho scan . --exclude-dirs .git,node_modules,build,dist,tmp
```

## System Requirements

- **No dependencies**: Single binary with everything included
- **Cross-platform**: Linux, macOS, Windows support
- **Permissions**: Requires read access to target directories
- **Memory**: Minimal memory usage, processes files individually

## Common Issues

### Permission Denied
```bash
# Make binary executable (Linux/macOS)
chmod +x codeecho
```

### Large Repositories
```bash
# For very large repos, exclude build directories
codeecho scan . --exclude-dirs .git,node_modules,target,build,dist,vendor
```

### Binary Files
Binary files are automatically excluded from scans. Only text files matching the included extensions are processed.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Support

For issues, questions, or contributions:
- **GitHub Issues**: [Report bugs or request features](https://github.com/opskraken/code-echo/issues)
- **Discussions**: [Community discussions](https://github.com/opskraken/code-echo/discussions)

---

*CodeEcho CLI - Making your repositories AI-ready*

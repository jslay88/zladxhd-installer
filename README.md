# ZLADXHD Installer

Automated installer for Zelda: Link's Awakening DX HD on Linux with Proton.

## Features

- Automatic protontricks installation
- Game archive download with caching and checksum verification
- Steam non-Steam game configuration
- Proton/Wine prefix setup
- .NET runtime installation via protontricks
- HD patcher download and execution

## Installation

### Arch Linux (AUR)

```bash
# Prebuilt binary (recommended)
yay -S zladxhd-installer-bin

# Build from source (latest git)
yay -S zladxhd-installer-git
```

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/jslay88/zladxhd-installer/releases):

```bash
# Download and extract
curl -sL https://github.com/jslay88/zladxhd-installer/releases/latest/download/zladxhd-installer-linux-amd64.tar.gz | tar xzv

# Make executable and move to PATH
chmod +x zladxhd-installer-linux-amd64
sudo mv zladxhd-installer-linux-amd64 /usr/local/bin/zladxhd-installer
```

For ARM64 systems, replace `amd64` with `arm64`.

### Build from Source

```bash
go install github.com/jslay88/zladxhd-installer/cmd/zladxhd-installer@latest
```

## Usage

```bash
# Interactive mode (will prompt for archive if not cached)
zladxhd-installer

# From local archive
zladxhd-installer --archive ~/Downloads/ZLADXHD.zip

# From URL (will cache in ~/.local/share/zladxhd-installer/cache/)
zladxhd-installer --archive https://example.com/ZLADXHD.zip

# With all options
zladxhd-installer \
  --archive ~/Downloads/ZLADXHD.zip \
  --install-dir ~/.local/share/Steam/steamapps/common/ZLADXHD \
  --proton "Proton 10.0" \
  --no-backup
```

## Options

| Flag | Description |
|------|-------------|
| `--archive, -a` | Path or URL to game archive (uses cache if not provided) |
| `--install-dir, -d` | Installation directory (default: `~/.local/share/Steam/steamapps/common/ZLADXHD`) |
| `--proton, -p` | Proton version to use (default: `Proton 10.0`) |
| `--no-backup` | Skip Steam backup prompt |
| `--backup` | Force Steam backup without prompt |

## Requirements

- Linux with Steam installed
- Go 1.25+ (for building from source)

## License

MIT

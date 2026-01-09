# AUR Package

This directory contains the PKGBUILD template for publishing to the [Arch User Repository (AUR)](https://aur.archlinux.org/).

## Package Name

`zladxhd-installer-bin` - Binary package (pre-compiled)

## Required GitHub Secrets

To enable automatic AUR publishing on release, configure these secrets in your GitHub repository:

| Secret | Description |
|--------|-------------|
| `AUR_USERNAME` | Your AUR username |
| `AUR_EMAIL` | Your email address for AUR commits |
| `AUR_SSH_PRIVATE_KEY` | SSH private key for AUR authentication |

## Manual Installation

Users can install from AUR using an AUR helper:

```bash
# Using yay
yay -S zladxhd-installer-bin

# Using paru
paru -S zladxhd-installer-bin
```

## PKGBUILD Template

The `PKGBUILD` file contains placeholders that are automatically replaced during the GitHub Actions release workflow:

- `VERSION_PLACEHOLDER` → Release version (e.g., `1.0.0`)
- `SHA256_AMD64_PLACEHOLDER` → SHA256 checksum of the amd64 binary
- `SHA256_ARM64_PLACEHOLDER` → SHA256 checksum of the arm64 binary

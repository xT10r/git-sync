# Git-Sync Service

## Prerequisites

- **Go 1.23+**: Main programming language
- **Docker**: For containerization
- **Make**: For build automation (optional on Windows)

## Documentation

- [Description (English)](./docs/readme.en.md)
- [Описание (Русский)](./docs/readme.ru.md)

## Building

### With Make (Linux/macOS)

```bash
# Build using Docker (recommended)
make build

# Build locally
make build-local

# Build Windows executable
make build-windows
```

### Without Make (Windows)

```powershell
# Build Windows executable
.\scripts\build-windows.ps1
```

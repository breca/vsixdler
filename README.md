# vsixdler

Download VS Code extensions from the marketplace and bundle them into a single release archive. Useful for offline installation, archival, and distribution.

**TL;DR:** Fork this repo, create a branch, edit `extensions.yaml` with the extensions you want, and push. A weekly GitHub Actions workflow will automatically download your extensions and publish them as release assets. You can also trigger it manually from the Actions tab.

## Setup

**Requirements:** Go 1.21+

```sh
git clone https://github.com/breca/vsixdler.git
cd vsixdler
go build -o vsixdler .
```

## Configuration

Create an `extensions.yaml` in the project root:

```yaml
extensions:
  - id: "ms-python.python"                    # latest, universal
  - id: "golang.go"
    version: "0.42.1"                          # pinned version
  - id: "ms-vscode.cpptools"
    platforms: [linux-x64, darwin-arm64]        # platform-specific
```

**Supported platforms:** `win32-x64`, `win32-arm64`, `linux-x64`, `linux-arm64`, `linux-armhf`, `alpine-x64`, `alpine-arm64`, `darwin-x64`, `darwin-arm64`, `web`

## Usage

```sh
# Preview what would be downloaded
./vsixdler download --dry-run

# Download to ./vsix/
./vsixdler download

# Custom output directory and config
./vsixdler download -c my-extensions.yaml -o ./output

# Parallel downloads (default: 4)
./vsixdler download -j 8

# Verbose logging
./vsixdler download -v
```

Downloaded files follow the naming convention:
```
ms-python.python-2024.17.vsix              # universal
ms-vscode.cpptools-1.22.5@linux-x64.vsix   # platform-specific
```

## GitHub Actions

The included workflow (`.github/workflows/sync-extensions.yaml`) runs weekly on Sundays at 04:00 UTC and can be triggered manually. It downloads all extensions and uploads them as GitHub release assets.

Each branch runs independently with its own `extensions.yaml`, so multiple users can maintain separate extension lists on different branches. Releases are tagged `{branch}/sync-{date}-{run}` to avoid collisions.

To use it, push your branch with an `extensions.yaml` and either wait for the schedule or trigger the workflow manually from the Actions tab.

# cleardir

Purge empty directories and dispensable files.

## Features

- Delete empty directories.
- Delete dispensable files such as `.DS_Store`.
- Prompt first and dry-mode: See what could or will be deleted before confirming.
- Max depth: Let's not dig too deep.

## Usage

```sh
# Clear the current directory.
cleardir

# Clear a specific directory.
cleardir /some/other/path

# Just display clearable items.
cleardir --dry-mode

# Trust me, I'm an engineer.
cleardir -y
```

## Installation

Below you'll find the recommended ways to install cleardir.

Alternatively, you can download cleardir from the [Releases](https://github.com/echocrow/cleardir/releases) page, or refer to [Development](#development) to build it yourself.

### macOS
Via [Homebrew](https://brew.sh/):
```sh
# Install:
brew install echocrow/tap/cleardir
# Update:
brew upgrade echocrow/tap/cleardir
```

## Configuration

By default, cleardir will not delete any files. To mark certain files safe for deletion, simply add the full filenames (including any extension) to the cleardir config file separated by a newline, e.g.
```
.DS_Store
temp.txt
.ephemeral
```

To get the path to the config file, run `cleardir --config '?'`.

macOS likes to generate `.DS_Store` files. To get rid of them with cleardir, you can add them to the config file with this command:
```sh
echo '.DS_Store' >> "$(cleardir --config '?')"
```

cleardir also accepts a custom config file path via `-c`/`--config`.

For more information and options, see `-h`/`--help`.

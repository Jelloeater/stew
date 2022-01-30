<p align="center">
  <img width=30% height=auto src="https://github.com/marwanhawari/stew/raw/main/assets/stew.png" alt="stew icon"/>
</p>

<h3 align="center">stew</h3>
<p align="center">
  An independent package manager for compiled binaries.
</p>
<p align="center">
  <img src="https://github.com/marwanhawari/stew/actions/workflows/test.yml/badge.svg" alt="build status"/>
  <img src="https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg" alt="code of conduct"/>
  <img src="https://img.shields.io/github/license/marwanhawari/stew?color=blue" alt="license"/>
</p>


# Features
* Easily distribute binaries across teams and private repositories.
* Get the latest releases ahead of other package managers.
* Rapidly browse, install, and experiment with different projects.
* Install binaries from GitHub releases or directly from URLs
* Isolated `~/.stew/` directory.
* No need for `sudo`.
* Portable [`Stewfile`](https://github.com/marwanhawari/stew/blob/main/examples/Stewfile) with optional pinned versioning.

![demo](https://github.com/marwanhawari/stew/raw/main/assets/demo.gif)

# Installation
Stew supports Linux, macOS, and Windows:
```
curl -fsSL https://raw.githubusercontent.com/marwanhawari/stew/main/install.sh | sh
```

# Usage
```sh
# Install from GitHub releases
stew install junegunn/fzf              # Install the latest release
stew install junegunn/fzf@v0.27.1      # Install a specific, tagged version
stew install junefunn/fzf sharkdp/fd   # Install multiple binaries in a single command

# Install directly from a URL
stew install https://github.com/cli/cli/releases/download/v2.4.0/gh_2.4.0_macOS_amd64.tar.gz

# Install from an Stewfile
stew install Stewfile

# Browse github releases and assets with a terminal UI
stew browse sharkdp/hyperfine

# Upgrade a binary to its latest version (only for binaries from GitHub releases)
stew upgrade rg           # Upgrade using the name of the binary directly
stew upgrade --all        # Upgrade all binaries

# Uninstall a binary
stew uninstall rg         # Uninstall using the name of the binary directly
stew uninstall --all      # Uninstall all binaries

# List installed binaries
stew list                              # Print to console
stew list > Stewfile                   # Create an Stewfile without pinned tags
stew list --tags > Stewfile            # Pin tags
stew list --tags --assets > Stewfile   # Pin tags and assets

```

# FAQ
### Why couldn't `stew` automatically find any binaries for X repo?
The repo probably uses an unconventional naming scheme for their binaries. You can always manually select the release asset.

### I've installed `stew` but the command is still not found.
The `stew` [install script](https://github.com/marwanhawari/stew/blob/main/install.sh) attempts to add `~/.stew/bin` to `PATH` in your `.zshrc` or `.bashrc` file. You will also need to start a new terminal session for the changes to take effect. Make sure that `~/.stew/bin` is in your `PATH` environment variable.

### Will `stew` work with private GitHub repositories?
Yes, `stew` will automatically detect if you have a `GITHUB_TOKEN` environment variable and allow you to access binaries from your private repositories.

### How do I uninstall `stew`?
Simply run `rm -rf $HOME/.stew/` and optionally remove this line
```
export PATH="$HOME/.stew/bin:$PATH"
```
from your `.zshrc` or `.bashrc` file.

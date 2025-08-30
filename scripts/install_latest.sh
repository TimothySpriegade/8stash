#!/bin/sh

# --- Configuration ---
OWNER="timothyspriegade"
REPO="8stash"
BINARY_NAME="8stash"
# -------------------
set -e
set -u

require() {
    if ! command -v "$1" > /dev/null 2>&1; then
        echo "Error: '$1' is not installed. Please install it first."
        exit 1
    fi
}

require "curl"
require "tar"
require "grep"

# Detect OS and Architecture
os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

case "$os" in
    linux)
        os="linux"
        ;;
    darwin)
        os="darwin"
        ;;
    *)
        echo "Error: Unsupported operating system '$os'. Only Linux and macOS are supported."
        exit 1
        ;;
esac

case "$arch" in
    x86_64 | amd64)
        arch="amd64"
        ;;
    arm64 | aarch64)
        arch="arm64"
        ;;
    i386 | i686)
        arch="386"
        ;;
    *)
        echo "Error: Unsupported architecture '$arch'."
        exit 1
        ;;
esac

# Fetch the latest release version from GitHub API
echo "Fetching the latest release version..."
latest_tag=$(curl -s "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$latest_tag" ]; then
    echo "Error: Could not fetch the latest release tag. Please check the repository details."
    exit 1
fi

version=$(echo "$latest_tag" | sed 's/v//') # Remove 'v' prefix if it exists
echo "Latest version is $version"

# Construct download URL and filenames
tar_file="${BINARY_NAME}_${version}_${os}_${arch}.tar.gz"
checksum_file="${BINARY_NAME}_${version}_checksums.txt"
download_url="https://github.com/${OWNER}/${REPO}/releases/download/${latest_tag}/${tar_file}"
checksum_url="https://github.com/${OWNER}/${REPO}/releases/download/${latest_tag}/${checksum_file}"

# Download the release archive and checksums file
echo "Downloading $tar_file..."
curl -sSL -o "/tmp/$tar_file" "$download_url"
echo "Downloading checksums..."
curl -sSL -o "/tmp/$checksum_file" "$checksum_url"

# Verify the checksum
echo "Verifying checksum..."
cd /tmp
if command -v "sha256sum" > /dev/null 2>&1; then
    # Linux
    grep "$tar_file" "$checksum_file" | sha256sum -c -
else
    # macOS
    expected_checksum=$(grep "$tar_file" "$checksum_file" | awk '{print $1}')
    calculated_checksum=$(shasum -a 256 "$tar_file" | awk '{print $1}')
    if [ "$expected_checksum" != "$calculated_checksum" ]; then
        echo "Error: Checksum verification failed."
        exit 1
    fi
fi
echo "Checksum verified."
cd - > /dev/null

# Extract the binary and install it
echo "Installing $BINARY_NAME..."
tar -xzf "/tmp/$tar_file" -C /tmp "$BINARY_NAME"

# Determine install location
install_dir="/usr/local/bin"
if [ ! -w "$install_dir" ]; then
    echo "Write permission to $install_dir is required."
    sudo install -m 755 "/tmp/$BINARY_NAME" "$install_dir/$BINARY_NAME"
else
    install -m 755 "/tmp/$BINARY_NAME" "$install_dir/$BINARY_NAME"
fi

# Clean up temporary files
rm -f "/tmp/$tar_file" "/tmp/$checksum_file" "/tmp/$BINARY_NAME"

echo "$BINARY_NAME was installed successfully to $install_dir/$BINARY_NAME"
echo "You can now run '$BINARY_NAME' from your terminal."
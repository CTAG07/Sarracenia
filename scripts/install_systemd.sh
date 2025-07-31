#!/bin/bash

# This script installs Sarracenia as a systemd service.
# It must be run with root privileges (e.g., using sudo).

set -e # Exit immediately if a command exits with a non-zero status.

# Configuration
SARRACENIA_USER="sarracenia"
SARRACENIA_GROUP="sarracenia"
INSTALL_DIR="/opt/sarracenia"
SERVICE_FILE_PATH="/etc/systemd/system/sarracenia.service"

# Check for root
if [ "$(id -u)" -ne 0 ]; then
  echo "This script must be run as root. Please use sudo." >&2
  exit 1
fi

echo "- Sarracenia Systemd Installer"

# Build
echo "Building the Sarracenia binary..."
echo "Choose your build type:"
echo "  1) Native Go (Recommended, no CGO dependencies)"
echo "  2) CGO with go-sqlite3 (Slightly higher performance, requires GCC)"

read -p "Enter choice [1-2]: " choice

VERSION=$(git describe --tags --abbrev=0)
COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS="-X 'main.Version=${VERSION}' -X 'main.Commit=${COMMIT}' -X 'main.BuildDate=${BUILD_DATE}'"

if [ "$choice" == "2" ]; then
    echo "Building with CGO enabled..."
    CGO_ENABLED=1 go build -ldflags="${LDFLAGS} -s -w" -o ./sarracenia ./cmd/main/
else
    echo "Building with native Go SQLite driver..."
    CGO_ENABLED=0 go build -ldflags="${LDFLAGS} -s -w" -o ./sarracenia ./cmd/main/
fi

echo "Build complete."

# User setup
if ! getent group "$SARRACENIA_GROUP" >/dev/null; then
    echo "Creating group '$SARRACENIA_GROUP'..."
    groupadd --system "$SARRACENIA_GROUP"
else
    echo "Group '$SARRACENIA_GROUP' already exists."
fi

if ! id "$SARRACENIA_USER" >/dev/null 2>&1; then
    echo "Creating user '$SARRACENIA_USER'..."
    useradd --system --no-create-home --gid "$SARRACENIA_GROUP" "$SARRACENIA_USER"
else
    echo "User '$SARRACENIA_USER' already exists."
fi

# Move example files to install dir
echo "Installing files to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
cp ./sarracenia "$INSTALL_DIR/sarracenia"
# Copy the 'example' directory which contains default config, dashboard, templates, etc.
# The contents of 'example' will be placed directly into INSTALL_DIR.
cp -r ./example/. "$INSTALL_DIR/"

echo "Setting permissions..."
chown -R "$SARRACENIA_USER":"$SARRACENIA_GROUP" "$INSTALL_DIR"
chmod 750 "$INSTALL_DIR"
chmod +x "$INSTALL_DIR/sarracenia"

# Service file install
echo "Installing systemd service file..."
cat > "$SERVICE_FILE_PATH" <<EOF
[Unit]
Description=Sarracenia Anti-Scraper Tarpit
After=network.target

[Service]
Type=simple
User=${SARRACENIA_USER}
Group=${SARRACENIA_GROUP}
WorkingDirectory=${INSTALL_DIR}
ExecStart=${INSTALL_DIR}/sarracenia
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# Systemd Service Install
echo "Reloading systemd daemon..."
systemctl daemon-reload

echo "Enabling Sarracenia to start on boot..."
systemctl enable sarracenia.service

echo "Starting Sarracenia service..."
systemctl start sarracenia.service

echo "- Installation Complete!"
echo "You can check the status with: systemctl status sarracenia.service"
echo "Logs can be viewed with:       journalctl -u sarracenia.service -f"
#!/bin/bash
# Script to install Playwright driver and browsers in the container

echo "Installing Playwright driver and browsers..."

# Install the Playwright CLI tool
go install github.com/playwright-community/playwright-go/cmd/playwright@latest

# Install the browsers (chromium is sufficient for testing)
echo "Installing Chromium browser..."
/go/bin/playwright install chromium

# Check installation
echo "Verifying installation..."
/go/bin/playwright --version

echo "Playwright installation complete!"
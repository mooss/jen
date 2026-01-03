#!/bin/bash
# Setup script for ArXiv Procedural Generation Papers Downloader

echo "Setting up ArXiv Procedural Generation Papers Downloader..."

# Check if uv is installed
if ! command -v uv &> /dev/null; then
    echo "uv not found"
    exit 1
fi

# Create virtual environment
echo "Creating virtual environment..."
uv venv

# Install dependencies
echo "Installing dependencies..."
uv pip install -r requirements.txt

echo "Setup complete!"
echo ""
echo "To run the downloader:"
echo "  source .venv/bin/activate"
echo "  python download_procgen_papers.py"
echo ""
echo "For help:"
echo "  python download_procgen_papers.py --help"

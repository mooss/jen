# ArXiv Procedural Generation Papers Downloader

A Python tool to search and download all arXiv papers related to procedural generation, generative algorithms, and related topics.

## Features

- Searches arXiv using multiple procedural generation-related terms
- Downloads PDFs of all matching papers
- Saves metadata and creates summary reports
- Concurrent downloads for faster processing
- Rate limiting to respect arXiv's API

## Installation

### Using uv (Recommended)

```bash
# Install uv if you don't have it
curl -LsSf https://astral.sh/uv/install.sh | sh

# Clone or download this repository
cd vibecoded/arxiv-procgen

# Create virtual environment and install dependencies
uv venv
uv pip install -r requirements.txt
```

### Using pip

```bash
pip install -r requirements.txt
```

## Usage

### Basic Usage

Download all procedural generation papers:
```bash
python download_procgen_papers.py
```

### Advanced Usage

```bash
# Download to custom directory
python download_procgen_papers.py --output-dir my_papers

# Limit number of papers
python download_procgen_papers.py --max-papers 100

# Only search and save metadata (skip PDF downloads)
python download_procgen_papers.py --skip-download
```

### Command Line Options

- `--output-dir, -o`: Directory to save downloaded papers (default: `procgen_papers`)
- `--max-papers, -m`: Maximum number of papers to download
- `--skip-download`: Only search and save metadata, skip downloading PDFs

## Output

The script creates several files in the output directory:

- `*.pdf`: Downloaded paper PDFs (named by arXiv ID)
- `papers_metadata.json`: JSON file with metadata for all papers
- `summary_report.txt`: Summary report with statistics and sample titles

## Search Terms

The script searches for papers using these procedural generation-related terms:

- procedural generation
- procedural content generation
- procedural modeling
- generative algorithms
- L-system
- fractal generation
- PCG (procedural game content)
- and many more...

## Rate Limiting

The script includes rate limiting to respect arXiv's API limits:
- 3-second delay between search queries
- Maximum 5 concurrent downloads

## Dependencies

- `requests`: For HTTP requests
- `feedparser`: For parsing arXiv RSS feeds
- `tqdm`: For progress bars

## License

This project is open source. Feel free to modify and distribute.

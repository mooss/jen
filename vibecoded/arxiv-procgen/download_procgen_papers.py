#!/usr/bin/env python3
"""
ArXiv Procedural Generation Papers Downloader

Search for and download all arXiv papers related to procedural generation.
"""

import os
import sys
import time
import requests
import feedparser
from pathlib import Path
from typing import List, Dict, Optional
import argparse
from tqdm import tqdm

# Constants
ARXIV_API_BASE = "http://export.arxiv.org/api/query"
MAX_RESULTS_PER_QUERY = 1000
DOWNLOAD_DIR = "procgen_papers"
MAX_WORKERS = 5
RATE_LIMIT_DELAY = 3 # Seconds between API calls, https://info.arxiv.org/help/api/tou.html#rate-limits.

# Search terms related to procedural generation
SEARCH_TERMS = [
    "procedural generation",
    "procedural content generation",
    "procedural modeling",
    "procedural synthesis",
    "generative algorithms",
    "generative systems",
    "algorithmic generation",
    "parametric generation",
    "rule-based generation",
    "grammar-based generation",
    "L-system",
    "fractal generation",
    "noise-based generation",
    "stochastic generation",
    "computer-generated content",
    "automated content creation",
    "generative art",
    "generative design",
    "generative architecture",
    "generative terrain",
    "generative textures",
    "generative music",
    "generative storytelling",
    "PCG",
    "procedural game content",
    "noise function",
]

def search_arxiv_papers(query: str, max_results: int = MAX_RESULTS_PER_QUERY) -> List[Dict]:
    """Search arXiv for papers matching the query."""
    params = {
        'search_query': f'all:"{query}"',
        'start': 0,
        'max_results': max_results,
        'sortBy': 'submittedDate',
        'sortOrder': 'descending'
    }

    try:
        response = requests.get(ARXIV_API_BASE, params=params)
        response.raise_for_status()

        feed = feedparser.parse(response.content)
        papers = []

        for entry in feed.entries:
            paper = {
                'title': entry.title,
                'authors': [author.name for author in entry.authors],
                'summary': entry.summary,
                'published': entry.published,
                'updated': entry.updated,
                'id': entry.id.split('/')[-1],  # Extract arXiv ID
                'pdf_url': entry.link.replace('abs', 'pdf') + '.pdf',
                'categories': [tag.term for tag in entry.tags]
            }
            papers.append(paper)

        return papers

    except requests.RequestException as e:
        print(f"Error searching arXiv for '{query}': {e}")
        return []

def download_paper(paper: Dict, download_dir: Path) -> bool:
    """Download a single paper PDF."""
    arxiv_id = paper['id']
    pdf_url = paper['pdf_url']

    # Create filename from arXiv ID
    filename = f"{arxiv_id}.pdf"
    filepath = download_dir / filename

    if filepath.exists():
        return True # Already downloaded

    # Make sure to sleep everytime a download is attempted
    time.sleep(RATE_LIMIT_DELAY)

    try:
        response = requests.get(pdf_url, stream=True)
        response.raise_for_status()

        # Save the PDF
        with open(filepath, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)

        return True

    except requests.RequestException as e:
        print(f"Error downloading {arxiv_id}: {e}")
        return False

def get_all_papers() -> List[Dict]:
    """Get all procedural generation papers from arXiv."""
    all_papers = []
    seen_ids = set()

    print("Searching arXiv for procedural generation papers...")

    for term in tqdm(SEARCH_TERMS, desc="Search terms"):
        papers = search_arxiv_papers(term)

        for paper in papers:
            if paper['id'] not in seen_ids:
                seen_ids.add(paper['id'])
                all_papers.append(paper)

        # Rate limiting
        time.sleep(RATE_LIMIT_DELAY)

    return all_papers

def download_all_papers(papers: List[Dict], download_dir: Path) -> None:
    """Download all papers with progress tracking."""
    print(f"\nDownloading {len(papers)} papers...")

    successful = 0
    failed = 0

    for paper in tqdm(papers, desc="Downloading"):
        success = download_paper(paper, download_dir)
        if success:
            successful += 1
        else:
            failed += 1

    print(f"\nDownload complete: {successful} successful, {failed} failed")

def save_paper_metadata(papers: List[Dict], download_dir: Path) -> None:
    """Save paper metadata as JSON for reference."""
    import json
    
    metadata_file = download_dir / "papers_metadata.json"
    
    # Convert papers to JSON-serializable format
    json_papers = []
    for paper in papers:
        json_paper = paper.copy()
        json_paper['authors'] = ', '.join(paper['authors'])
        json_paper['categories'] = ', '.join(paper['categories'])
        json_papers.append(json_paper)
    
    with open(metadata_file, 'w', encoding='utf-8') as f:
        json.dump(json_papers, f, indent=2, ensure_ascii=False)
    
    print(f"Saved metadata for {len(papers)} papers to {metadata_file}")

def create_summary_report(papers: List[Dict], download_dir: Path) -> None:
    """Create a summary report of the downloaded papers."""
    report_file = download_dir / "summary_report.txt"

    with open(report_file, 'w', encoding='utf-8') as f:
        f.write("ArXiv Procedural Generation Papers - Summary Report\n")
        f.write("=" * 50 + "\n\n")
        f.write(f"Total papers found: {len(papers)}\n")
        f.write(f"Search date: {time.strftime('%Y-%m-%d %H:%M:%S')}\n\n")

        # Count papers by year
        years = {}
        for paper in papers:
            year = paper['published'][:4]
            years[year] = years.get(year, 0) + 1

        f.write("Papers by year:\n")
        for year in sorted(years.keys(), reverse=True):
            f.write(f"  {year}: {years[year]}\n")

        f.write("\nTop categories:\n")
        # Count categories
        categories = {}
        for paper in papers:
            for cat in paper['categories']:
                if cat.startswith('cs.'):  # Computer science categories
                    categories[cat] = categories.get(cat, 0) + 1

        sorted_cats = sorted(categories.items(), key=lambda x: x[1], reverse=True)
        for cat, count in sorted_cats[:10]:
            f.write(f"  {cat}: {count}\n")

        f.write("\nSample paper titles:\n")
        for i, paper in enumerate(papers[:10]):
            f.write(f"  {i+1}. {paper['title']}\n")

    print(f"Summary report saved to {report_file}")

def main():
    parser = argparse.ArgumentParser(description="Download arXiv papers about procedural generation")
    parser.add_argument("--output-dir", "-o", default=DOWNLOAD_DIR,
                       help="Directory to save downloaded papers (default: procgen_papers)")
    parser.add_argument("--max-papers", "-m", type=int, default=None,
                       help="Maximum number of papers to download (default: all)")
    parser.add_argument("--skip-download", action="store_true",
                       help="Only search and save metadata, skip downloading PDFs")

    args = parser.parse_args()

    # Create download directory
    download_dir = Path(args.output_dir)
    download_dir.mkdir(exist_ok=True)

    # Search for papers
    papers = get_all_papers()

    if not papers:
        print("No papers found!")
        return

    print(f"\nFound {len(papers)} unique papers")

    # Limit papers if requested
    if args.max_papers and len(papers) > args.max_papers:
        papers = papers[:args.max_papers]
        print(f"Limited to {len(papers)} papers")

    # Save metadata
    save_paper_metadata(papers, download_dir)

    # Create summary report
    create_summary_report(papers, download_dir)

    # Download papers unless skipped
    if not args.skip_download:
        download_all_papers(papers, download_dir)
    else:
        print("Skipping download as requested")

if __name__ == "__main__":
    main()

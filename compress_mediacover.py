#!/usr/bin/env python3
"""
compress_mediacover.py — Radarr/Sonarr MediaCover image compressor

Finds all fanart.jpg and poster.jpg under MediaCover directories and
recompresses them in-place, optionally keeping backups.

Usage:
    python3 compress_mediacover.py [OPTIONS] PATH [PATH ...]

Examples:
    # Dry-run on all arr stacks
    python3 compress_mediacover.py --dry-run /etc/komodo/stacks/arr/*/

    # Compress with defaults (fanart → 85% quality, poster → 80%)
    python3 compress_mediacover.py /etc/komodo/stacks/arr/radarr /etc/komodo/stacks/arr/sonarr

    # Aggressive compression with backup
    python3 compress_mediacover.py --fanart-quality 60 --poster-quality 65 --backup /etc/komodo/stacks/arr/radarr/

    # Compress only fanart, resize fanart max width to 1920px
    python3 compress_mediacover.py --no-posters --fanart-max-width 1920 /etc/komodo/stacks/arr/radarranime/

    # JSON output for scripting
    python3 compress_mediacover.py --json /etc/komodo/stacks/arr/radarr/
"""

import argparse
import json
import os
import shutil
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional

try:
    from PIL import Image
except ImportError:
    print("ERROR: Pillow not installed. Run: pip install Pillow --break-system-packages", file=sys.stderr)
    sys.exit(1)


# ── defaults ──────────────────────────────────────────────────────────────────

FANART_QUALITY   = 75   # JPEG quality for fanart.jpg
POSTER_QUALITY   = 72   # JPEG quality for poster.jpg
FANART_MAX_WIDTH = 1920 # resize fanart if wider than this (0 = no resize)
POSTER_MAX_WIDTH = 1000 # resize poster if wider than this (0 = no resize)
MIN_SAVING_KB    = 50   # skip file if saving would be less than this

TARGET_FILES = {"fanart.jpg", "poster.jpg"}


# ── data ──────────────────────────────────────────────────────────────────────

@dataclass
class FileResult:
    path: str
    original_bytes: int
    new_bytes: int
    skipped: bool = False
    skip_reason: str = ""
    error: str = ""

    @property
    def saved_bytes(self) -> int:
        return self.original_bytes - self.new_bytes

    @property
    def saved_pct(self) -> float:
        if self.original_bytes == 0:
            return 0.0
        return self.saved_bytes / self.original_bytes * 100


@dataclass
class Summary:
    results: list[FileResult] = field(default_factory=list)

    @property
    def processed(self):  return [r for r in self.results if not r.skipped and not r.error]
    @property
    def skipped(self):    return [r for r in self.results if r.skipped]
    @property
    def errors(self):     return [r for r in self.results if r.error]
    @property
    def total_saved(self): return sum(r.saved_bytes for r in self.processed)
    @property
    def total_original(self): return sum(r.original_bytes for r in self.processed)


# ── core ──────────────────────────────────────────────────────────────────────

def find_targets(roots: list[Path]) -> list[Path]:
    """Walk each root and collect fanart.jpg / poster.jpg files."""
    found = []
    for root in roots:
        if not root.exists():
            print(f"  [warn] path not found: {root}", file=sys.stderr)
            continue
        for path in root.rglob("*"):
            if path.is_file() and path.name in TARGET_FILES:
                found.append(path)
    return sorted(found)


def compress_image(
    src: Path,
    quality: int,
    max_width: int,
    dry_run: bool,
    backup: bool,
) -> FileResult:
    orig_size = src.stat().st_size
    result = FileResult(path=str(src), original_bytes=orig_size, new_bytes=orig_size)

    try:
        img = Image.open(src)
        orig_w, orig_h = img.size

        # Convert to RGB if needed (handles palette/RGBA)
        if img.mode not in ("RGB", "L"):
            img = img.convert("RGB")

        # Resize if over max_width
        new_w, new_h = orig_w, orig_h
        if max_width and orig_w > max_width:
            ratio  = max_width / orig_w
            new_w  = max_width
            new_h  = int(orig_h * ratio)
            if not dry_run:
                img = img.resize((new_w, new_h), Image.LANCZOS)

        if dry_run:
            # Estimate compressed size via in-memory buffer
            import io
            buf = io.BytesIO()
            tmp = img.copy()
            if max_width and orig_w > max_width:
                tmp = tmp.resize((new_w, new_h), Image.LANCZOS)
            tmp.save(buf, "JPEG", quality=quality, optimize=True, progressive=True)
            est_size = buf.tell()
            saving = orig_size - est_size
            if saving < MIN_SAVING_KB * 1024:
                result.skipped    = True
                result.skip_reason = f"saving < {MIN_SAVING_KB} KB (est. {saving/1024:.0f} KB)"
                result.new_bytes   = orig_size
            else:
                result.new_bytes = est_size
            return result

        # Real write
        saving = orig_size - _estimate_size(img, quality)
        if saving < MIN_SAVING_KB * 1024:
            result.skipped    = True
            result.skip_reason = f"saving < {MIN_SAVING_KB} KB"
            return result

        if backup:
            shutil.copy2(src, src.with_suffix(".bak.jpg"))

        img.save(src, "JPEG", quality=quality, optimize=True, progressive=True)
        result.new_bytes = src.stat().st_size

    except Exception as exc:
        result.error = str(exc)

    return result


def _estimate_size(img: Image.Image, quality: int) -> int:
    import io
    buf = io.BytesIO()
    img.save(buf, "JPEG", quality=quality, optimize=True, progressive=True)
    return buf.tell()


# ── display ───────────────────────────────────────────────────────────────────

def fmt_bytes(b: int) -> str:
    if b < 1024:        return f"{b} B"
    if b < 1024**2:     return f"{b/1024:.1f} KB"
    if b < 1024**3:     return f"{b/1024**2:.1f} MB"
    return f"{b/1024**3:.2f} GB"

GREEN  = "\033[92m"
YELLOW = "\033[93m"
RED    = "\033[91m"
CYAN   = "\033[96m"
DIM    = "\033[2m"
RESET  = "\033[0m"
BOLD   = "\033[1m"

def color_arrow(pct: float) -> str:
    if pct >= 50: return GREEN
    if pct >= 20: return CYAN
    return DIM

def print_result(r: FileResult, dry_run: bool):
    label = Path(r.path).name
    rel   = str(Path(r.path).parent).replace("/etc/komodo/stacks/arr/", "…/")
    if r.error:
        print(f"  {RED}✗{RESET}  {rel}/{label}  {RED}{r.error}{RESET}")
    elif r.skipped:
        print(f"  {DIM}–{RESET}  {rel}/{label}  {DIM}skipped ({r.skip_reason}){RESET}")
    else:
        arrow = color_arrow(r.saved_pct)
        tag   = "[dry]" if dry_run else ""
        print(
            f"  {GREEN}✓{RESET}  {rel}/{label}  "
            f"{fmt_bytes(r.original_bytes)} → {arrow}{fmt_bytes(r.new_bytes)}{RESET}  "
            f"{arrow}−{r.saved_pct:.1f}%{RESET}  {DIM}{tag}{RESET}"
        )


# ── main ──────────────────────────────────────────────────────────────────────

def parse_args():
    p = argparse.ArgumentParser(
        description="Compress Radarr/Sonarr MediaCover fanart and poster images.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )
    p.add_argument("paths", nargs="+", help="Root paths to search (e.g. /etc/komodo/stacks/arr/radarr/)")
    p.add_argument("--dry-run",  action="store_true", help="Estimate savings without modifying any file")
    p.add_argument("--backup",   action="store_true", help="Keep .bak.jpg alongside each modified file")
    p.add_argument("--json",     action="store_true", help="Output results as JSON (implies --quiet)")
    p.add_argument("--quiet", "-q", action="store_true", help="Suppress per-file output")

    g = p.add_argument_group("quality")
    g.add_argument("--fanart-quality", type=int, default=FANART_QUALITY, metavar="Q",
                   help=f"JPEG quality for fanart.jpg (default: {FANART_QUALITY})")
    g.add_argument("--poster-quality", type=int, default=POSTER_QUALITY, metavar="Q",
                   help=f"JPEG quality for poster.jpg (default: {POSTER_QUALITY})")
    g.add_argument("--fanart-max-width", type=int, default=FANART_MAX_WIDTH, metavar="PX",
                   help=f"Resize fanart if wider than PX (default: {FANART_MAX_WIDTH}, 0=off)")
    g.add_argument("--poster-max-width", type=int, default=POSTER_MAX_WIDTH, metavar="PX",
                   help=f"Resize poster if wider than PX (default: {POSTER_MAX_WIDTH}, 0=off)")
    g.add_argument("--min-saving", type=int, default=MIN_SAVING_KB, metavar="KB",
                   help=f"Skip file if saving is less than KB (default: {MIN_SAVING_KB})")

    g2 = p.add_argument_group("targets")
    g2.add_argument("--no-fanart",  action="store_true", help="Skip fanart.jpg files")
    g2.add_argument("--no-posters", action="store_true", help="Skip poster.jpg files")

    return p.parse_args()


def main():
    args   = parse_args()
    roots  = [Path(p) for p in args.paths]
    quiet  = args.quiet or args.json
    prefix = f"{YELLOW}[DRY-RUN]{RESET} " if args.dry_run else ""

    targets = find_targets(roots)
    if not targets:
        print("No fanart.jpg / poster.jpg files found.")
        sys.exit(0)

    # Filter based on --no-fanart / --no-posters
    if args.no_fanart:
        targets = [t for t in targets if t.name != "fanart.jpg"]
    if args.no_posters:
        targets = [t for t in targets if t.name != "poster.jpg"]

    if not quiet:
        print(f"\n{prefix}{BOLD}compress_mediacover{RESET}  {len(targets)} files found\n")

    summary = Summary()
    for path in targets:
        is_fanart = path.name == "fanart.jpg"
        quality   = args.fanart_quality if is_fanart else args.poster_quality
        max_width = args.fanart_max_width if is_fanart else args.poster_max_width

        result = compress_image(
            src       = path,
            quality   = quality,
            max_width = max_width,
            dry_run   = args.dry_run,
            backup    = args.backup,
        )
        summary.results.append(result)
        if not quiet:
            print_result(result, args.dry_run)

    # ── summary ──
    if args.json:
        out = {
            "dry_run": args.dry_run,
            "processed": len(summary.processed),
            "skipped": len(summary.skipped),
            "errors": len(summary.errors),
            "total_original_bytes": summary.total_original,
            "total_saved_bytes": summary.total_saved,
            "files": [
                {
                    "path": r.path,
                    "original_bytes": r.original_bytes,
                    "new_bytes": r.new_bytes,
                    "saved_bytes": r.saved_bytes,
                    "saved_pct": round(r.saved_pct, 2),
                    "skipped": r.skipped,
                    "skip_reason": r.skip_reason,
                    "error": r.error,
                }
                for r in summary.results
            ],
        }
        print(json.dumps(out, indent=2))
        return

    if not quiet:
        print()
        print(f"  {'─'*60}")
        n = len(summary.processed)
        if n:
            orig = summary.total_original
            saved = summary.total_saved
            pct   = saved / orig * 100 if orig else 0
            tag   = f"{YELLOW}(estimated){RESET}" if args.dry_run else ""
            print(
                f"  {BOLD}saved  {GREEN}{fmt_bytes(saved)}{RESET} / {fmt_bytes(orig)}"
                f"  ({GREEN}−{pct:.1f}%{RESET})  across {n} files  {tag}"
            )
        if summary.skipped:
            print(f"  {DIM}skipped  {len(summary.skipped)} files (below min saving threshold){RESET}")
        if summary.errors:
            print(f"  {RED}errors   {len(summary.errors)} files{RESET}")
        print()


if __name__ == "__main__":
    main()

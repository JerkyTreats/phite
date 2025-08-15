#!/usr/bin/env python3
import argparse
import logging
import os
import sys
import glob
import threading
import time
from pathlib import Path
from typing import List, Tuple

import duckdb

LOGGER = logging.getLogger("gwas.inspect_enrichment_state")
logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")


def _fmt_bytes(n: int) -> str:
    # Simple size formatter (avoid external deps)
    for unit in ["B", "KB", "MB", "GB", "TB"]:
        if n < 1024:
            return f"{n:.0f} {unit}" if unit == "B" else f"{n:.1f} {unit}"
        n /= 1024
    return f"{n:.1f} PB"


def list_parquet_files(parquet_dir: Path) -> List[Tuple[Path, int]]:
    files = []
    for p in sorted(parquet_dir.glob("alleles_chr*.parquet")):
        try:
            files.append((p, p.stat().st_size))
        except FileNotFoundError:
            continue
    return files


def count_parquet_rows_with_progress(parquet_files: List[Path], heartbeat_interval: int = 10, quiet: bool = False) -> int:
    """Count rows across Parquet files, emitting per-file progress and a periodic heartbeat."""
    if not parquet_files:
        return 0
    con = duckdb.connect()
    total = 0
    files_done = 0
    nfiles = len(parquet_files)
    start_ts = time.time()
    stop_evt = threading.Event()

    def heartbeat():
        while not stop_evt.wait(max(1, heartbeat_interval)):
            elapsed = time.time() - start_ts
            LOGGER.info(
                "Heartbeat: counted %d/%d parquet files, %s rows so far (elapsed %.1fs)",
                files_done,
                nfiles,
                f"{total:,}",
                elapsed,
            )

    hb = threading.Thread(target=heartbeat, name="inspect-heartbeat", daemon=True)
    hb.start()

    try:
        for idx, p in enumerate(parquet_files, start=1):
            if not quiet:
                LOGGER.info("[%d/%d] Counting rows in %s", idx, nfiles, p.name)
            rows = int(con.execute("SELECT COUNT(*) FROM parquet_scan(?)", [str(p)]).fetchone()[0])
            total += rows
            files_done = idx
            if not quiet:
                LOGGER.info("[%d/%d] %s: %s rows (cumulative %s)", idx, nfiles, p.name, f"{rows:,}", f"{total:,}")
    finally:
        stop_evt.set()
        hb.join(timeout=1)
        con.close()
    elapsed = time.time() - start_ts
    LOGGER.info("Parquet counting complete: %s rows across %d files in %.1fs", f"{total:,}", nfiles, elapsed)
    return total


def inspect_duckdb(db_path: Path) -> dict:
    result = {
        "exists": db_path.exists(),
        "tables": [],
        "views": [],
        "snp_alleles_count": None,
        "chr_distribution": [],
        "indexes": [],
    }
    if not db_path.exists():
        return result

    con = duckdb.connect(str(db_path))
    try:
        result["tables"] = [r[0] for r in con.execute("PRAGMA show_tables;").fetchall()]
        try:
            result["views"] = [r[0] for r in con.execute("PRAGMA show_views;").fetchall()]
        except Exception:
            result["views"] = []
        try:
            result["indexes"] = [
                tuple(r) for r in con.execute("PRAGMA show_indexes;").fetchall()
            ]
        except Exception:
            pass

        if "snp_alleles" in result["tables"]:
            result["snp_alleles_count"] = int(
                con.execute("SELECT COUNT(*) FROM snp_alleles;").fetchone()[0]
            )
            try:
                # Order numerically for autosomes, then others
                dist = con.execute(
                    """
                    SELECT chr, COUNT(*) AS c
                    FROM snp_alleles
                    GROUP BY 1
                    ORDER BY
                      CASE WHEN chr ~ '^[0-9]+$' THEN CAST(chr AS INT) ELSE 99 END,
                      chr
                    """
                ).fetchall()
                result["chr_distribution"] = [(r[0], int(r[1])) for r in dist]
            except Exception:
                pass
    finally:
        con.close()

    return result


def main():
    parser = argparse.ArgumentParser(
        description="Inspect GWAS allele enrichment state: Parquet cache vs DuckDB contents"
    )
    parser.add_argument(
        "--duckdb",
        type=Path,
        default=Path("gwas/gwas.duckdb"),
        help="Path to DuckDB database file",
    )
    parser.add_argument(
        "--parquet_dir",
        type=Path,
        default=Path("gwas/gwas/parquet/alleles"),
        help="Directory containing allele Parquet files",
    )
    parser.add_argument(
        "--heartbeat-interval",
        type=int,
        default=10,
        help="Seconds between heartbeat logs during counting (default: 10)",
    )
    parser.add_argument(
        "--quiet",
        action="store_true",
        help="Reduce per-file logging during counting",
    )
    args = parser.parse_args()

    print("=== Parquet Cache ===")
    if not args.parquet_dir.exists():
        print(f"Parquet directory not found: {args.parquet_dir}")
        sys.exit(1)

    parquet_list = list_parquet_files(args.parquet_dir)
    print(f"Directory: {args.parquet_dir}")
    print(f"Files found: {len(parquet_list)}")
    if parquet_list:
        # Show a compact listing
        for p, sz in parquet_list:
            print(f"  {p.name:20s}  { _fmt_bytes(sz):>8s}")
    parquet_paths = [p for p, _ in parquet_list]
    try:
        parquet_rows = count_parquet_rows_with_progress(parquet_paths, heartbeat_interval=args.heartbeat_interval, quiet=args.quiet)
        print(f"Total rows across Parquet: {parquet_rows}")
    except Exception as e:
        print(f"Error counting Parquet rows: {e}")
        parquet_rows = None

    print("\n=== DuckDB State ===")
    db_path = args.duckdb
    print(f"DuckDB path: {db_path}  exists={db_path.exists()}")
    state = inspect_duckdb(db_path)

    if not state["exists"]:
        print("DuckDB database file not found. Enrichment import likely not performed yet.")
    else:
        print(f"Tables: {sorted(state['tables'])}")
        print(f"Views: {sorted(state['views'])}")
        if state["snp_alleles_count"] is None:
            print("Table 'snp_alleles' not found.")
        else:
            print(f"snp_alleles row count: {state['snp_alleles_count']}")
            if state["chr_distribution"]:
                print("Chromosome distribution (chr, count):")
                for chr_name, cnt in state["chr_distribution"]:
                    print(f"  {chr_name:>3s}: {cnt}")

    print("\n=== Assessment ===")
    if state["snp_alleles_count"] is None:
        print("No imported allele data found in DuckDB. You can run enrich_with_alleles.py to import.")
    else:
        if parquet_rows is not None:
            diff = parquet_rows - state["snp_alleles_count"]
            print(
                f"Parquet total vs DuckDB table: {parquet_rows} vs {state['snp_alleles_count']} (delta={diff})"
            )
        else:
            print("Could not compute Parquet total to compare.")
        print(
            "The import script uses INSERT OR REPLACE with a PRIMARY KEY (rsid) and CREATE IF NOT EXISTS,\n"
            "so re-running enrich_with_alleles.py is safe and idempotent. It will resume by re-reading\n"
            "cached Parquet files and upserting into DuckDB."
        )


if __name__ == "__main__":
    main()

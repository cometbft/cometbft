#!/usr/bin/env python3
# run with: python3 scripts/bench/p2p_bench.py --concurrency 2
import argparse
import json
import os
import sqlite3
import subprocess
import threading
import time
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

DISCOVERY_TESTCASE_RE = r"^\s*testCase:\s*([^\s]+)\s*$"
DISCOVERY_RUN_RE = r"TestBench/([^\s]+)"
INPUT_RE = r"\[INPUT\]:\s*(\{.*\})"
OUTPUT_RE = r"\[OUTPUT\]:\s*(\{.*\})"
DEFAULT_CONCURRENCY = 2
TEST_TIMEOUT = "60s"

@dataclass
class BenchResult:
    index: int
    total: int
    test_name: str
    config: dict[str, Any] | None
    output: dict[str, Any] | None
    logs: str
    time_taken_sec: float
    return_code: int

def repo_root() -> Path:
    return Path(__file__).resolve().parents[2]

def init_db(db_path: Path) -> None:
    db_path.parent.mkdir(parents=True, exist_ok=True)
    with sqlite3.connect(db_path) as conn:
        conn.execute(
            """
            CREATE TABLE IF NOT EXISTS benchmarks (
                id INTEGER PRIMARY KEY,
                added_at TEXT,
                test_name TEXT,
                config TEXT,
                logs TEXT,
                output TEXT,
                time_taken_sec REAL
            )
            """
        )
        conn.commit()


def discover_tests(root: Path) -> list[str]:
    env = os.environ.copy()
    env["P2P_BENCH_TEST"] = "matrix-only"
    cmd = [
        "go",
        "test",
        "-run",
        "TestBench",
        "./lp2p/...",
        "-count=1",
        "-v",
    ]
    proc = subprocess.run(
        cmd,
        cwd=root,
        capture_output=True,
        text=True,
        env=env,
        check=False,
    )
    combined = f"{proc.stdout}\n{proc.stderr}"
    return parse_discovered_tests(combined)


def parse_discovered_tests(output: str) -> list[str]:
    import re

    tests: list[str] = []
    seen: set[str] = set()

    for line in output.splitlines():
        case_match = re.match(DISCOVERY_TESTCASE_RE, line)
        if case_match:
            name = case_match.group(1)
            if name not in seen:
                tests.append(name)
                seen.add(name)
            continue

        for run_match in re.finditer(DISCOVERY_RUN_RE, line):
            name = run_match.group(1)
            if name not in seen:
                tests.append(name)
                seen.add(name)

    return tests


def parse_json_marker(logs: str, marker_re: str) -> dict[str, Any] | None:
    import re

    parsed: dict[str, Any] | None = None
    for line in logs.splitlines():
        m = re.search(marker_re, line)
        if not m:
            continue
        raw = m.group(1).strip()
        try:
            candidate = json.loads(raw)
            if isinstance(candidate, dict):
                parsed = candidate
        except json.JSONDecodeError:
            continue
    return parsed


def run_single_test(root: Path, test_name: str, index: int, total: int) -> BenchResult:
    import re

    env = os.environ.copy()
    env["P2P_BENCH_TEST"] = "1"

    run_pattern = "^TestBench/" + re.escape(test_name) + "$"
    cmd = [ "go", "test", "-run", run_pattern, "-count=1", "-timeout", TEST_TIMEOUT, "./lp2p/...", "-v" ]

    started = time.perf_counter()
    proc = subprocess.run(
        cmd,
        cwd=root,
        capture_output=True,
        text=True,
        env=env,
        check=False,
    )
    elapsed = time.perf_counter() - started
    logs = f"{proc.stdout}\n{proc.stderr}"

    return BenchResult(
        index=index,
        total=total,
        test_name=test_name,
        config=parse_json_marker(logs, INPUT_RE),
        output=parse_json_marker(logs, OUTPUT_RE),
        logs=logs,
        time_taken_sec=elapsed,
        return_code=proc.returncode,
    )


def insert_result(conn: sqlite3.Connection, result: BenchResult) -> None:
    added_at = datetime.now(timezone.utc).isoformat()
    conn.execute(
        """
        INSERT INTO benchmarks (added_at, test_name, config, logs, output, time_taken_sec)
        VALUES (?, ?, ?, ?, ?, ?)
        """,
        (
            added_at,
            result.test_name,
            json.dumps(result.config) if result.config is not None else None,
            result.logs,
            json.dumps(result.output) if result.output is not None else None,
            result.time_taken_sec,
        ),
    )
    conn.commit()


def main() -> int:
    from concurrent.futures import FIRST_COMPLETED, ThreadPoolExecutor, wait

    parser = argparse.ArgumentParser(description="Run P2P benchmark matrix and save results.")
    parser.add_argument("--concurrency", type=int, default=DEFAULT_CONCURRENCY)
    args = parser.parse_args()

    if args.concurrency < 1:
        raise SystemExit("--concurrency must be >= 1")

    root = repo_root()
    db_path = root / "scripts" / "bench" / "p2p_bench.sqlite"
    init_db(db_path)

    tests = discover_tests(root)
    if not tests:
        print("discovered 0 tests")
        return 1

    total_tests = len(tests)
    print(f"discovered {total_tests} tests")

    print_lock = threading.Lock()
    results: list[BenchResult] = []

    with sqlite3.connect(db_path) as conn, ThreadPoolExecutor(
        max_workers=args.concurrency
    ) as executor:
        test_iter = iter(tests)
        active_futures: dict[Any, tuple[int, str]] = {}
        next_index = 1

        def submit_next() -> bool:
            nonlocal next_index
            try:
                test_name = next(test_iter)
            except StopIteration:
                return False

            with print_lock:
                print(f"[{next_index}/{total_tests}] running test {test_name}")
            active_futures[
                executor.submit(run_single_test, root, test_name, next_index, total_tests)
            ] = (next_index, test_name)
            next_index += 1
            return True

        for _ in range(min(args.concurrency, len(tests))):
            submit_next()

        while active_futures:
            done, _ = wait(active_futures.keys(), return_when=FIRST_COMPLETED)
            for future in done:
                index, test_name = active_futures.pop(future)
                result = future.result()
                results.append(result)
                insert_result(conn, result)
                with print_lock:
                    suffix = ""
                    if result.return_code != 0:
                        suffix = f" (exit={result.return_code})"
                    print(
                        f"[{index}/{total_tests}] done {test_name}. "
                        f"time taken: {result.time_taken_sec:.1f}s{suffix}"
                    )
                submit_next()

    return 0


if __name__ == "__main__":
    raise SystemExit(main())

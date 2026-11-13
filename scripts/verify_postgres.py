#!/usr/bin/env python3
"""Verify a converted dump against the original, using an existing test container."""

import json
import re
import subprocess
import sys

from convert_mysql_dump import SOURCE, values

container = sys.argv[1]
database = sys.argv[2] if len(sys.argv) > 2 else "clims_validation"


def query(sql):
    result = subprocess.run(
        [
            "docker",
            "exec",
            "-i",
            container,
            "psql",
            "-U",
            "postgres",
            "-d",
            database,
            "-At",
            "-v",
            "ON_ERROR_STOP=1",
        ],
        input=sql,
        text=True,
        capture_output=True,
        check=True,
    )
    return result.stdout.strip()


def rows(data):
    """Parse literal tuples from the translated SQL without executing them."""
    row = []
    i = 0
    while i < len(data):
        c = data[i]
        if c in " (,\n\r":
            i += 1
        elif c == ")":
            yield row
            row = []
            i += 1
        elif c == "'":
            i += 1
            value = []
            while i < len(data):
                if data[i] == "'":
                    if i + 1 < len(data) and data[i + 1] == "'":
                        value.append("'")
                        i += 2
                        continue
                    i += 1
                    break
                value.append(data[i])
                i += 1
            row.append("".join(value))
        else:
            end = i
            while end < len(data) and data[end] not in ",)":
                end += 1
            literal = data[i:end]
            row.append(None if literal == "NULL" else int(literal))
            i = end


source = SOURCE.read_text()
expected = {}
for match in re.finditer(r"CREATE TABLE `([^`]+)` \(\n(.*?)\n\) ENGINE=[^;]+;", source, re.S):
    table, body = match.groups()
    columns = re.findall(r"^  `([^`]+)` (\w+)", body, re.M)
    expected[table] = (columns, [])
for line in source.splitlines():
    if not line.startswith("INSERT INTO "):
        continue
    match = re.fullmatch(r"INSERT INTO `([^`]+)` VALUES (.*);", line)
    table, data = match.groups()
    columns, records = expected[table]
    for record in rows(values(data)[0]):
        assert len(record) == len(columns), f"Column count: {table}"
        records.append(
            {
                name: value.replace(" ", "T") if kind == "timestamp" and value else value
                for (name, kind), value in zip(columns, record)
            }
        )


def canonical(row):
    return json.dumps(row, sort_keys=True, ensure_ascii=False)


count = 0
for table, (columns, records) in expected.items():
    actual = json.loads(query(f"SELECT COALESCE(json_agg(t), '[]'::json) FROM \"{table}\" t;"))
    # Compare every value without printing personal data on failure.
    assert sorted(map(canonical, records)) == sorted(map(canonical, actual)), (
        f"Data mismatch: {table}"
    )
    count += len(records)
    if any(name == "id" and kind in ("bigint", "int") for name, kind in columns):
        next_id = int(query(f"SELECT nextval(pg_get_serial_sequence('{table}', 'id'));"))
        assert next_id > max((r["id"] for r in records), default=0), f"Invalid sequence: {table}"
print(
    f"PASS: all values match for {count} rows across {len(expected)} tables; identity sequences are safe."
)

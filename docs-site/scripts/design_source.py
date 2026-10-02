"""Verify and materialize the independent harness design reference."""

import hashlib
import json
from pathlib import Path
import tarfile

SITE = Path(__file__).resolve().parents[1]
SOURCE = SITE / "design/harness"


def reference_files():
    manifest = json.loads((SOURCE / "source.json").read_text())
    archive_path = SOURCE / "reference.tar.gz"
    if hashlib.sha256(archive_path.read_bytes()).hexdigest() != manifest["archiveSha256"]:
        raise ValueError("Harness reference archive hash mismatch")
    with tarfile.open(archive_path) as archive:
        files = {}
        for row in manifest["files"]:
            member = archive.extractfile(row["path"])
            if member is None:
                raise ValueError(f"Missing reference file: {row['path']}")
            content = member.read()
            if hashlib.sha256(content).hexdigest() != row["sha256"]:
                raise ValueError(f"Reference hash mismatch: {row['path']}")
            files[row["path"]] = content
    replacements = json.loads((SOURCE / "copy.json").read_text())
    for path, mapping in replacements.items():
        content = files[path].decode()
        for before, after in sorted(mapping.items(), key=lambda row: len(row[0]), reverse=True):
            if before not in content:
                raise ValueError(f"Missing text replacement in {path}: {before}")
            content = content.replace(before, after)
        files[path] = content.encode()
    return manifest, files


def verify():
    manifest, files = reference_files()
    for row in manifest["files"]:
        if row.get("referenceOnly"):
            continue
        path = row["path"]
        if (SITE / path).read_bytes() != files[path]:
            raise ValueError(f"Harness design drift: {path}; refresh from the pinned source or change product copy in design/harness/copy.json")
    print(f"Verified {len(files)} harness design inputs at {manifest['commit']}.")


if __name__ == "__main__":
    verify()

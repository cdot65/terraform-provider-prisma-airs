#!/usr/bin/env python3
"""Run the complete directional example against a live Runtime tenant.

Requires PANW_MGMT_* credentials and a built provider. Creates only a unique
disposable profile, borrows an existing DLP profile read-only, and removes all
owned revisions in finally. Raw Terraform/API receipts stay in a private temp
directory; stdout contains only scoped verification results.
"""

import base64
import copy
import datetime
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import urllib.parse
import urllib.request
import uuid


ROOT = Path(__file__).resolve().parents[1]
ADDRESS = "prisma-airs_runtime_security_profile.directional"
BASE = os.environ.get("PANW_MGMT_ENDPOINT", "https://api.sase.paloaltonetworks.com/aisec")


def main():
    os.umask(0o077)
    work = Path(tempfile.mkdtemp(prefix="airs-directional-live-"))
    name = "tf-live-directional-" + uuid.uuid4().hex[:12]
    print("Evidence directory:", work, flush=True)
    print("UTC:", datetime.datetime.now(datetime.timezone.utc).isoformat(), flush=True)
    print("Owned profile:", name, flush=True)
    print("Provider binary SHA256:", hashlib.sha256(
        (ROOT / "terraform-provider-prisma-airs").read_bytes()).hexdigest(), flush=True)
    form = urllib.parse.urlencode({
        "grant_type": "client_credentials",
        "scope": "tsg_id:" + os.environ["PANW_MGMT_TSG_ID"],
    }).encode()
    credentials = base64.b64encode((os.environ["PANW_MGMT_CLIENT_ID"] + ":" +
                                    os.environ["PANW_MGMT_CLIENT_SECRET"]).encode()).decode()
    request = urllib.request.Request(
        os.environ.get("PANW_MGMT_TOKEN_ENDPOINT", "https://auth.apps.paloaltonetworks.com/oauth2/access_token"),
        form, {"Authorization": "Basic " + credentials,
               "Content-Type": "application/x-www-form-urlencoded"})
    with urllib.request.urlopen(request, timeout=60) as response:
        token = json.load(response)["access_token"]

    def api(path, method="GET", body=None):
        request = urllib.request.Request(BASE.rstrip("/") + path,
                                         None if body is None else json.dumps(body).encode(),
                                         {"Authorization": "Bearer " + token,
                                          "Content-Type": "application/json"}, method=method)
        with urllib.request.urlopen(request, timeout=60) as response:
            raw = response.read()
        return json.loads(raw) if method != "DELETE" else None

    def revisions():
        result, offset = [], 0
        while True:
            page = api("/v1/mgmt/profiles/tsg/" + os.environ["PANW_MGMT_TSG_ID"] +
                       f"?offset={offset}&limit=100&latest=false")
            result.extend(p for p in page["ai_profiles"] if p["profile_name"] == name)
            next_offset = page.get("next_offset", 0)
            if next_offset > offset:
                offset = next_offset
            elif len(page["ai_profiles"]) >= 100:
                offset += len(page["ai_profiles"])
            else:
                return result

    def receipt(label):
        profile = max(revisions(), key=lambda p: p["revision"])
        (work / (label + "-api.json")).write_text(json.dumps(profile, indent=2) + "\n")
        print(f"{label}: revision={profile['revision']} profile_id={profile['profile_id']}", flush=True)
        return profile

    env = os.environ.copy()
    config = work / "terraform.tfrc"
    config.write_text('provider_installation {\n  dev_overrides {\n    "cdot65/prisma-airs" = ' +
                      json.dumps(str(ROOT)) + '\n  }\n  direct {}\n}\n')
    env["TF_CLI_CONFIG_FILE"] = str(config)
    env["TF_IN_AUTOMATION"] = "1"

    def terraform(label, *args, allowed=(0,)):
        result = subprocess.run(["terraform", *args], cwd=work, env=env,
                                stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
        (work / (label + ".log")).write_text(result.stdout)
        print(f"{label}: exit={result.returncode}", flush=True)
        for line in result.stdout.splitlines():
            if re.match(r"^(Plan:|Apply complete!|Destroy complete!|No changes\.|Import successful!)", line):
                print(line, flush=True)
        if result.returncode not in allowed:
            raise RuntimeError(f"{label} failed; inspect private receipt {work / (label + '.log')}")
        return result

    assert not revisions(), "Fixture name already exists; refusing ownership"
    dlp = api("/v1/mgmt/dlpprofiles?offset=0&limit=100")["dlp_profiles"]
    borrowed = next(p for p in dlp if p.get("name") and p.get("version"))
    env["TF_VAR_profile_name"] = name
    env["TF_VAR_dlp_profile_name"] = borrowed["name"]
    env["TF_VAR_dlp_profile_version"] = borrowed["version"]
    print("Borrowed DLP reference: read-only=true", flush=True)
    source = (ROOT / "examples/directional-security-profile/main.tf").read_text()
    (work / "main.tf").write_text(source)
    known_ids = set()
    try:
        terraform("validate", "validate", "-no-color")
        terraform("create", "apply", "-auto-approve", "-no-color")
        initial = receipt("created")
        known_ids.add(initial["profile_id"])
        entry = initial["policy"]["ai-security-profiles"][0]
        shared_flags = entry["model-configuration"]
        assert shared_flags["mask-data-in-storage"] is False
        assert shared_flags["enable-full-conversation-inspection"] is False
        directions = entry["content-type-configurations"]
        assert set(directions) == {"prompt", "response", "tool-call", "tool-response"}
        for direction in ("prompt", "response", "tool-call"):
            assert directions[direction]["data-protection"]["data-leak-detection"]["mask-data-inline"] is True
        assert "mask-data-inline" not in directions["tool-response"]["data-protection"]["data-leak-detection"]
        print("Presence: shared=false inline=true tool-response-inline=omitted", flush=True)
        terraform("refresh-stable", "plan", "-detailed-exitcode", "-no-color")
        terraform("state-remove", "state", "rm", ADDRESS)
        terraform("import", "import", "-no-color", ADDRESS, name)
        terraform("import-stable", "plan", "-detailed-exitcode", "-no-color")

        # Change only response toxicity. Compare live JSON for every other direction.
        response_start = source.index("      response {")
        response_end = source.index("      tool_call {")
        changed = source[:response_start] + source[response_start:response_end].replace(
            'high     = "medium"', 'high     = "high"') + source[response_end:]
        (work / "main.tf").write_text(changed)
        terraform("response-update", "apply", "-auto-approve", "-no-color")
        updated = receipt("response-updated")
        known_ids.add(updated["profile_id"])
        before = initial["policy"]["ai-security-profiles"][0]["content-type-configurations"]
        after = updated["policy"]["ai-security-profiles"][0]["content-type-configurations"]
        for direction in ("prompt", "tool-call", "tool-response"):
            assert before[direction] == after[direction], direction + " changed unexpectedly"
        assert before["response"] != after["response"], "Response update was not persisted"
        print("Response-only update: prompt/tool-call/tool-response preserved=true", flush=True)
        terraform("response-stable", "plan", "-detailed-exitcode", "-no-color")

        changed = changed.replace("max_inline_latency    = 5", "max_inline_latency    = 6")
        (work / "main.tf").write_text(changed)
        terraform("shared-update", "apply", "-auto-approve", "-no-color")
        shared = receipt("shared-updated")
        known_ids.add(shared["profile_id"])
        assert after == shared["policy"]["ai-security-profiles"][0]["content-type-configurations"]
        print("Shared latency update: all four directions preserved=true", flush=True)

        detector = '''        model_protection {
          action   = "block"
          name     = "prompt-injection"
          severity = "medium"
        }
'''
        assert detector in changed
        changed = changed.replace(detector, "", 1)
        (work / "main.tf").write_text(changed)
        terraform("detector-remove", "apply", "-auto-approve", "-no-color")
        removed = receipt("detector-removed")
        known_ids.add(removed["profile_id"])
        policy = removed["policy"]["ai-security-profiles"][0]
        assert not any(d["name"] == "prompt-injection" for d in
                       policy["content-type-configurations"]["prompt"]["model-protection"])
        terraform("removal-stable", "plan", "-detailed-exitcode", "-no-color")
        print("Prompt detector removal: persisted=true", flush=True)

        # A scoped SDK-independent mutation proves the provider reports remote drift.
        drift = copy.deepcopy(removed["policy"])
        detectors = drift["ai-security-profiles"][0]["content-type-configurations"]["response"]["model-protection"]
        toxic = next(d for d in detectors if d["name"] == "toxic-content")
        toxic["severity-by-confidence"]["high"] = "low"
        api("/v1/mgmt/profile/uuid/" + removed["profile_id"], "PUT",
            {"profile_name": name, "policy": drift})
        drifted = receipt("remote-drift")
        known_ids.add(drifted["profile_id"])
        terraform("drift-plan", "plan", "-detailed-exitcode", "-no-color", allowed=(2,))
        print("Remote response severity drift: detected=true", flush=True)
        terraform("drift-repair", "apply", "-auto-approve", "-no-color")
        repaired = receipt("drift-repaired")
        known_ids.add(repaired["profile_id"])
        terraform("final-stable", "plan", "-detailed-exitcode", "-no-color")
        terraform("destroy", "destroy", "-auto-approve", "-no-color")
    finally:
        # Failure paths may create a revision before Terraform records its state.
        for profile in revisions():
            known_ids.add(profile["profile_id"])
            api("/v1/mgmt/profile/" + profile["profile_id"] +
                "/force?updated_by=terraform-live-example", "DELETE")
        assert not revisions(), "Owned profile revisions remain after cleanup"
        print(f"Cleanup: remaining_revisions=0 observed_revision_ids={len(known_ids)}", flush=True)
        hashes = {p.name: hashlib.sha256(p.read_bytes()).hexdigest()
                  for p in work.iterdir() if p.suffix == ".log" or p.name.endswith("-api.json")}
        (work / "receipt-sha256.json").write_text(json.dumps(hashes, indent=2) + "\n")


if __name__ == "__main__":
    main()

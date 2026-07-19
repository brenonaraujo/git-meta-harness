#!/usr/bin/env python3
# ============================================================================
# git-meta-harness — Check Feature Flow (Sensor 13, v1.13.0)
# ============================================================================
# Python companion to check-feature-flow.sh. Reads GitHub issues
# (JSON via stdin) and validates the feature flow:
#
#   type/feature  →  refined (domain-expert refinement comment)  →  ready
#   (architect DoD comment)  →  in-progress (builder)
#
# Categories detected:
#   - no_refined_label:    issue is type/feature but missing `refined` label
#   - no_ready_label:     issue is type/feature but missing `ready` label
#   - no_refinement_comment: no domain-expert refinement comment with ACs
#   - no_dod_comment:     no architect DoD comment
#   - refined_before_body: refinement comment posted before issue body
#                         (suggests domain-expert didn't read the body)
#   - dod_without_refined: DoD comment but no refinement (architect ran first)
#
# BLOCKING — exit 1 on any violation for type/feature issues.
# ============================================================================

import json
import os
import re
import sys

# Read issues JSON from stdin (one per line, gh output format)
# gh output can be either:
#   - single JSON object (when --json is used on `gh issue view`)
#   - JSON array (when --json is used on `gh issue list`)
raw = sys.stdin.read().strip()
if not raw:
    print("No issues fetched.")
    sys.exit(0)

try:
    parsed = json.loads(raw)
except json.JSONDecodeError as e:
    print(f"❌ Invalid JSON from stdin: {e}", file=sys.stderr)
    sys.exit(2)

if isinstance(parsed, dict):
    issues = [parsed]
elif isinstance(parsed, list):
    issues = parsed
else:
    print(f"❌ Unexpected JSON shape: {type(parsed)}", file=sys.stderr)
    sys.exit(2)

REPO = os.environ.get("REPO", "")

# Pattern matchers
AC_RE = re.compile(
    r"\b(AC[- ]?\d+|acceptance criteria|crit[ée]rios? de aceite|✅.*AC)\b",
    re.IGNORECASE,
)
EDGE_CASE_RE = re.compile(
    r"\b(edge case|casos? de borda|⚠️.*EC|EC[- ]?\d+)\b",
    re.IGNORECASE,
)
DOD_RE = re.compile(
    r"\b(definition of done|DoD|criteria de pronto|done criteria|🏛️.*DoD)\b",
    re.IGNORECASE,
)
PILAR_RE = re.compile(
    r"\b(pilar|pillar|🏛️\s*Pilar)\b",
    re.IGNORECASE,
)

# Persona author identifiers (matches login or bot suffix)
DOMAIN_EXPERT_RE = re.compile(r"domain-expert", re.IGNORECASE)
SOLUTIONS_ARCHITECT_RE = re.compile(r"(solutions-architect|architect)", re.IGNORECASE)


def get_label_names(labels):
    if not labels:
        return set()
    return {l.get("name", "") for l in labels}


def get_comments_list(issue):
    c = issue.get("comments")
    if isinstance(c, list):
        return c
    if isinstance(c, dict) and "nodes" in c:
        return c["nodes"]
    return []


def is_type_feature(labels):
    names = get_label_names(labels)
    return "type/feature" in names


def find_refinement(comments):
    """Returns the first comment that looks like a domain-expert
    refinement (has ACs + edge cases), or None."""
    for c in comments:
        body = c.get("body", "") or ""
        author = (c.get("author") or {}).get("login", "") or ""
        # Author can be the persona bot or a human domain-expert
        if not (DOMAIN_EXPERT_RE.search(author) or "domain-expert" in body.lower()):
            continue
        if AC_RE.search(body) and EDGE_CASE_RE.search(body):
            return c
    # If no domain-expert author match, look for AC+EC pattern
    # in any comment posted by a person (not the issue author)
    for c in comments:
        body = c.get("body", "") or ""
        if AC_RE.search(body) and EDGE_CASE_RE.search(body):
            return c
    return None


def find_dod(comments):
    """Returns the first comment that looks like an architect
    DoD (has DoD + pillars), or None."""
    for c in comments:
        body = c.get("body", "") or ""
        author = (c.get("author") or {}).get("login", "") or ""
        if not (SOLUTIONS_ARCHITECT_RE.search(author) or "architect" in body.lower()):
            continue
        if DOD_RE.search(body) and (PILAR_RE.search(body) or "pilar" in body.lower()):
            return c
    # Fallback: any comment with DoD + pillars
    for c in comments:
        body = c.get("body", "") or ""
        if DOD_RE.search(body) and (PILAR_RE.search(body) or "pilar" in body.lower()):
            return c
    return None


violations = []
checked = 0

for issue in issues:
    if issue.get("state") == "CLOSED":
        continue  # skip closed issues
    if not is_type_feature(issue.get("labels", [])):
        continue  # not a feature, not in scope

    checked += 1
    number = issue.get("number", "?")
    title = (issue.get("title") or "")[:60]
    labels = get_label_names(issue.get("labels", []))
    comments = get_comments_list(issue)

    # 1. Has `refined` label?
    if "refined" not in labels:
        violations.append({
            "issue": number,
            "title": title,
            "category": "no_refined_label",
            "fix": "Domain-expert must refine the story. See "
                   "harness/templates/comments/domain-expert-refinement.md",
        })

    # 2. Has `ready` label?
    if "ready" not in labels:
        violations.append({
            "issue": number,
            "title": title,
            "category": "no_ready_label",
            "fix": "Solutions-architect must define DoD. See "
                   "harness/templates/comments/solutions-architect-dod.md",
        })

    # 3. Has a refinement comment with ACs + edge cases?
    refinement = find_refinement(comments)
    if not refinement:
        violations.append({
            "issue": number,
            "title": title,
            "category": "no_refinement_comment",
            "fix": "Post refinement comment using the template (with ACs and edge cases).",
        })

    # 4. Has a DoD comment with pillars?
    dod = find_dod(comments)
    if not dod:
        violations.append({
            "issue": number,
            "title": title,
            "category": "no_dod_comment",
            "fix": "Post DoD comment using the template (with pillars and DoD items).",
        })

    # 5. If DoD exists but no refinement, that's architect running
    #    without domain context (architect should refine DoD, not
    #    do the domain work)
    if dod and not refinement:
        violations.append({
            "issue": number,
            "title": title,
            "category": "dod_without_refined",
            "fix": "Architect should run AFTER domain-expert. Get domain-expert to refine first.",
        })

if checked == 0:
    print("No open type/feature issues found.")
    sys.exit(0)

# Report
print(f"Checked {checked} type/feature issue(s).")
if not violations:
    print()
    print("OK: All type/feature issues have refined + ready + DoD.")
    sys.exit(0)

# Group by issue
by_issue = {}
for v in violations:
    by_issue.setdefault(v["issue"], []).append(v)

print()
print(f"BLOCKING: FEATURE FLOW VIOLATIONS (sensor 13, v1.13.0):")
print()
for num, issues_list in by_issue.items():
    title = issues_list[0]["title"]
    print(f"  Issue #{num}: {title}")
    for v in issues_list:
        print(f"    ❌ {v['category']}")
        print(f"       Fix: {v['fix']}")
    print()

print(f"Total: {len(violations)} violation(s) across {len(by_issue)} issue(s).")
print()
print("Recovery (per category):")
print("  no_refined_label:     ask domain-expert to refine (use template)")
print("  no_ready_label:       ask solutions-architect to define DoD (use template)")
print("  no_refinement_comment: post refinement comment with ACs + edge cases")
print("  no_dod_comment:       post DoD comment with pillars + DoD items")
print("  dod_without_refined:  run domain-expert first, then architect")
print()
print("Templates:")
print("  harness/templates/comments/domain-expert-refinement.md")
print("  harness/templates/comments/solutions-architect-dod.md")
print()
print("Run this sensor BEFORE moving labels to in-progress.")
print()

sys.exit(1)

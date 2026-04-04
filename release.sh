#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

# ── Determine next version ──────────────────────────────────────
latest=$(git tag --sort=-v:refname | grep -m1 '^v[0-9].*-beta\.')

if [[ -z "$latest" ]]; then
    echo "ERROR: no beta tag found"
    exit 1
fi

current="${latest#v}"
base="${current%-beta.*}"
num="${current##*-beta.}"
next_num=$((num + 1))
next="${base}-beta.${next_num}"
next_tag="v${next}"

echo "Current: ${current}"
echo "Next:    ${next}"

# ── Update npm/package.json ─────────────────────────────────────
pkg="npm/package.json"
if [[ ! -f "$pkg" ]]; then
    echo "ERROR: ${pkg} not found"
    exit 1
fi

tmp=$(mktemp)
sed "s/\"version\": \"[^\"]*\"/\"version\": \"${next}\"/" "$pkg" > "$tmp"
mv "$tmp" "$pkg"

echo "Updated ${pkg} → ${next}"

# ── Git commit, tag, push ───────────────────────────────────────
git add "$pkg"
git commit -m "chore: bump npm version to ${next}"
git tag "$next_tag"
echo "Tagged ${next_tag}"

BRANCH=$(git branch --show-current)
echo "Pushing ${BRANCH} + tag ${next_tag} ..."
git push origin "$BRANCH" --tags

echo ""
echo "Done! Pushed ${next_tag} — CI will handle npm publish."

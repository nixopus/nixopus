#!/usr/bin/env bash
set -euo pipefail

# Regroups a conventional-changelog file from per-release entries into monthly groups.
# Drops bullets whose normalized text duplicates an earlier bullet anywhere in the file
# (same rule as stripping parentheticals, lowercasing, collapsing spaces — first wins).
# Usage: ./regroup-changelog.sh [input] [output]

INPUT="${1:-CHANGELOG.md}"
OUTPUT="${2:-$INPUT}"
REPO_URL="https://github.com/nixopus/nixopus"

awk -v repo_url="$REPO_URL" '
function month_to_num(name, names, i) {
    split("January February March April May June July August September October November December", names, " ")
    for (i = 1; i <= 12; i++) if (names[i] == name) return sprintf("%02d", i)
    return ""
}

BEGIN {
    month_count = 0
    current_version = ""
    current_date = ""
    current_section = ""
    expect_preserved_gt = ""
    section_order[1] = "Features"
    section_order[2] = "Bug Fixes"
    section_order[3] = "Performance Improvements"
    section_count = 3
}

# Existing regrouped body: ## [March 2026](url) then optional "> ..." line
/^## \[/ {
    line = $0
    rb = index(line, "]")
    if (rb < 6) next
    label = substr(line, 5, rb - 5)
    nf = split(label, lp, " ")
    if (nf == 2 && lp[2] ~ /^[0-9]{4}$/) {
        mn = month_to_num(lp[1])
        if (mn != "") {
            year = lp[2]
            current_date = year "-" mn "-01"
            month_key = year "-" mn
            current_section = ""
            if (!(month_key in month_seen)) {
                month_seen[month_key] = 1
                month_count++
                month_keys[month_count] = month_key
            }
            is_preserved[month_key] = 1
            preserved_h2[month_key] = $0
            expect_preserved_gt = month_key
            next
        }
    }
}

/^> / && expect_preserved_gt != "" {
    preserved_gt[expect_preserved_gt] = $0
    expect_preserved_gt = ""
    next
}

/^#+ \[?[0-9]/ {
    # Parse both formats:
    #   # [version](url) (YYYY-MM-DD)   — linked
    #   # version (YYYY-MM-DD)           — bare
    line = $0
    if (index(line, "[") > 0 && match(line, /\[[0-9][^\]]+\]/)) {
        current_version = substr(line, RSTART + 1, RLENGTH - 2)
    } else {
        sub(/^#+ +/, "", line)
        sub(/ +\(.*/, "", line)
        current_version = line
    }

    if (match($0, /[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]/)) {
        current_date = substr($0, RSTART, RLENGTH)
    }
    current_section = ""

    month_key = substr(current_date, 1, 7)

    if (!(month_key in month_seen)) {
        month_seen[month_key] = 1
        month_count++
        month_keys[month_count] = month_key
    }

    n = ++version_count[month_key]
    versions[month_key, n] = current_version
    dates[month_key, n] = current_date
    expect_preserved_gt = ""
    next
}

/^### / {
    if (expect_preserved_gt != "") expect_preserved_gt = ""
    current_section = substr($0, 5)
    next
}

/^\* / && current_section != "" {
    month_key = substr(current_date, 1, 7)
    key = month_key SUBSEP current_section

    # Extract 7-char commit hash for dedup: pattern [abcdef0]
    commit_hash = ""
    if (match($0, /\[[a-f0-9][a-f0-9][a-f0-9][a-f0-9][a-f0-9][a-f0-9][a-f0-9]\]/)) {
        commit_hash = substr($0, RSTART + 1, 7)
    }

    # Build description for dedup (matches prior standalone dedupe pass)
    desc = $0
    gsub(/\([^)]*\)/, "", desc)
    desc = tolower(desc)
    gsub(/[[:space:]]+/, " ", desc)

    if (desc in seen_global) next
    if (commit_hash != "" && (month_key SUBSEP commit_hash) in seen_hash) next
    if ((month_key SUBSEP current_section SUBSEP desc) in seen_desc) next

    seen_global[desc] = 1
    if (commit_hash != "") seen_hash[month_key SUBSEP commit_hash] = 1
    seen_desc[month_key SUBSEP current_section SUBSEP desc] = 1

    n = ++item_count[key]
    items[key, n] = $0

    if (!((month_key SUBSEP current_section) in section_seen)) {
        section_seen[month_key SUBSEP current_section] = 1
        sn = ++section_count_per_month[month_key]
        sections_in_month[month_key, sn] = current_section
    }
}

function month_name(m) {
    split("January,February,March,April,May,June,July,August,September,October,November,December", names, ",")
    return names[int(m)]
}

END {
    print "# Changelog\n"
    print "All notable changes to [Nixopus](" repo_url ") are documented in this file.\n"
    print "This changelog is grouped by month. For the full commit history, see the [compare view on GitHub](" repo_url "/commits/master).\n"
    print "---\n"

    # Sort month keys descending
    n = month_count
    for (i = 1; i <= n; i++) sorted[i] = month_keys[i]
    for (i = 1; i < n; i++)
        for (j = i + 1; j <= n; j++)
            if (sorted[i] < sorted[j]) { tmp = sorted[i]; sorted[i] = sorted[j]; sorted[j] = tmp }

    for (mi = 1; mi <= n; mi++) {
        mk = sorted[mi]
        year = substr(mk, 1, 4)
        mon = substr(mk, 6, 2)
        label = month_name(mon) " " year

        vc = version_count[mk]

        if (is_preserved[mk] && vc == 0) {
            print preserved_h2[mk]
            if (mk in preserved_gt) print preserved_gt[mk]
            print ""
        } else {
            first_date = dates[mk, 1]; first_ver = versions[mk, 1]
            last_date  = dates[mk, 1]; last_ver  = versions[mk, 1]
            for (vi = 2; vi <= vc; vi++) {
                if (dates[mk, vi] < first_date) { first_date = dates[mk, vi]; first_ver = versions[mk, vi] }
                if (dates[mk, vi] > last_date)  { last_date  = dates[mk, vi]; last_ver  = versions[mk, vi] }
            }

            compare = repo_url "/compare/v" first_ver "...v" last_ver
            if (first_ver == last_ver)
                vrange = first_ver
            else
                vrange = first_ver " ... " last_ver

            plural = (vc == 1) ? "release" : "releases"

            printf "## [%s](%s)\n", label, compare
            printf "> `%s` (%d %s)\n\n", vrange, vc, plural
        }

        has_content = 0

        for (si = 1; si <= section_count; si++) {
            sec = section_order[si]
            key = mk SUBSEP sec
            ic = item_count[key]
            if (ic > 0) {
                has_content = 1
                printf "### %s\n\n", sec
                for (ii = 1; ii <= ic; ii++) print items[key, ii]
                print ""
            }
        }

        sc = section_count_per_month[mk]
        for (si = 1; si <= sc; si++) {
            sec = sections_in_month[mk, si]
            skip = 0
            for (oi = 1; oi <= section_count; oi++)
                if (section_order[oi] == sec) skip = 1
            if (skip) continue

            key = mk SUBSEP sec
            ic = item_count[key]
            if (ic > 0) {
                has_content = 1
                printf "### %s\n\n", sec
                for (ii = 1; ii <= ic; ii++) print items[key, ii]
                print ""
            }
        }

        if (!has_content) print "_No notable changes._\n"
        print ""
    }
}
' "$INPUT" > "${OUTPUT}.tmp"

mv "${OUTPUT}.tmp" "$OUTPUT"
echo "Regrouped changelog written to $OUTPUT"

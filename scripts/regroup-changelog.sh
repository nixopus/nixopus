#!/usr/bin/env bash
set -euo pipefail

# Regroups a conventional-changelog file from per-release entries into monthly groups.
# Usage: ./regroup-changelog.sh [input] [output]

INPUT="${1:-CHANGELOG.md}"
OUTPUT="${2:-$INPUT}"
REPO_URL="https://github.com/raghavyuva/nixopus"

awk -v repo_url="$REPO_URL" '
BEGIN {
    month_count = 0
    current_version = ""
    current_date = ""
    current_section = ""
    section_order[1] = "Features"
    section_order[2] = "Bug Fixes"
    section_order[3] = "Performance Improvements"
    section_count = 3
}

/^# \[/ {
    # Parse: # [version](url) (YYYY-MM-DD)
    line = $0
    gsub(/^# \[/, "", line)
    split(line, p1, /\]/)
    current_version = p1[1]

    if (match(line, /[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]/)) {
        current_date = substr(line, RSTART, RLENGTH)
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
    next
}

/^### / {
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

    # Build description for dedup
    desc = $0
    gsub(/\([^)]*\)/, "", desc)
    desc = tolower(desc)
    gsub(/[[:space:]]+/, " ", desc)

    if (commit_hash != "" && (month_key SUBSEP commit_hash) in seen_hash) next
    if ((month_key SUBSEP current_section SUBSEP desc) in seen_desc) next

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

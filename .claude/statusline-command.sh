#!/bin/bash
# Single-line "Instrument" statusline:
# [modes] ◆model effort │ dir@branch● ↑2 │ ctx dots + tokens + burn │ $cost · duration │ 5h % ⟳reset · 7d %

set -f  # disable globbing

input=$(cat)

if [ -z "$input" ]; then
    printf "Claude"
    exit 0
fi

# ===== ANSI colors =====
blue='\033[38;2;0;153;255m'
orange='\033[38;2;255;176;85m'
green='\033[38;2;0;160;0m'
cyan='\033[38;2;46;149;153m'
red='\033[38;2;255;85;85m'
yellow='\033[38;2;230;200;0m'
white='\033[38;2;220;220;220m'
magenta='\033[38;2;192;103;222m'
dim='\033[2m'
reset='\033[0m'

# ===== Helpers =====

format_tokens() {
    local num=$1
    if [ "$num" -ge 1000000 ]; then
        awk "BEGIN {printf \"%.1fm\", $num / 1000000}"
    elif [ "$num" -ge 1000 ]; then
        awk "BEGIN {printf \"%.0fk\", $num / 1000}"
    else
        printf "%d" "$num"
    fi
}

usage_color() {
    local pct=$1
    if [ "$pct" -ge 90 ]; then echo "$red"
    elif [ "$pct" -ge 70 ]; then echo "$orange"
    elif [ "$pct" -ge 50 ]; then echo "$yellow"
    else echo "$green"
    fi
}

progress_dots() {
    local pct=$1
    local filled=$(( pct / 20 ))
    [ "$filled" -gt 5 ] && filled=5
    local empty=$(( 5 - filled ))
    local bar_color
    bar_color=$(usage_color "$pct")
    local bar=""
    local i
    for (( i=0; i<filled; i++ )); do bar+="●"; done
    for (( i=0; i<empty; i++ )); do bar+="○"; done
    echo "${bar_color}${bar}${reset}"
}

# epoch -> "2h13m" / "11m" / "" if already past
fmt_eta() {
    local target=$1 now=$2 diff h m
    [ -z "$target" ] || [ "$target" = "0" ] && return
    diff=$(( target - now ))
    [ "$diff" -le 0 ] && return
    h=$(( diff / 3600 ))
    m=$(( (diff % 3600) / 60 ))
    if [ "$h" -gt 0 ]; then printf "%dh%02dm" "$h" "$m"; else printf "%dm" "$m"; fi
}

sep=" ${dim}·${reset} "
gsep=" ${dim}│${reset} "
mkdir -p /tmp/claude
now_epoch=$(date +%s)

# ===== Extract everything in one jq pass =====
# Booleans are mapped to "1"/"" so empty-string tests stay unambiguous.
eval "$(printf '%s' "$input" | jq -r '{
    session_id:   (.session_id // "nosession"),
    cwd:          (.cwd // ""),
    size:         (.context_window.context_window_size // 200000),
    in_tok:       (.context_window.current_usage.input_tokens // 0),
    cache_create: (.context_window.current_usage.cache_creation_input_tokens // 0),
    cache_read:   (.context_window.current_usage.cache_read_input_tokens // 0),
    json_pct:     (.context_window.used_percentage // -1),
    cost_usd:     (.cost.total_cost_usd // 0),
    duration_ms:  (.cost.total_duration_ms // 0),
    model_name:   (.model.display_name // ""),
    effort_level: (.effort.level // ""),
    fast_mode:    (if .fast_mode then "1" else "" end),
    no_think:     (if .thinking.enabled == false then "1" else "" end),
    style_name:   (if (.output_style.name // "default") == "default" then "" else .output_style.name end),
    vim_mode:     (.vim.mode // ""),
    agent_name:   (.agent.name // ""),
    worktree:     (.worktree.name // ""),
    h5_pct:       (.rate_limits.five_hour.used_percentage // -1),
    h5_reset:     (.rate_limits.five_hour.resets_at // 0),
    d7_pct:       (.rate_limits.seven_day.used_percentage // -1),
    d7_reset:     (.rate_limits.seven_day.resets_at // 0)
} | to_entries | map("\(.key)=\(.value | tostring | @sh)") | .[]')"

# Headroom feed — maximum-effort's pool pick and its scripts/usage.sh read ~/.claude/rate-limits.json.
if [ "$d7_pct" -ge 0 ] 2>/dev/null; then
    rl_tmp="${HOME}/.claude/rate-limits.json.$$"
    printf '%s' "$input" | jq -c --argjson ts "$now_epoch" '{ts: $ts, model: .model.display_name, effort: .effort.level, five_hour: .rate_limits.five_hour, seven_day: .rate_limits.seven_day}' > "$rl_tmp" 2>/dev/null \
        && mv -f "$rl_tmp" "${HOME}/.claude/rate-limits.json"
fi

current=$(( in_tok + cache_create + cache_read ))
used_tokens=$(format_tokens $current)
total_tokens=$(format_tokens $size)

# Trust the payload's percentage; fall back to computing it.
if [ "$json_pct" -ge 0 ] 2>/dev/null; then
    pct_used=$json_pct
elif [ "$size" -gt 0 ]; then
    pct_used=$(( current * 100 / size ))
else
    pct_used=0
fi

cost_str=""
if [ -n "$cost_usd" ] && [ "$cost_usd" != "0" ] && [ "$cost_usd" != "null" ]; then
    cost_str=$(printf '$%.2f' "$cost_usd")
fi

duration_str=""
if [ "$duration_ms" -gt 0 ] 2>/dev/null; then
    duration_s=$(( duration_ms / 1000 ))
    hours=$(( duration_s / 3600 ))
    mins=$(( (duration_s % 3600) / 60 ))
    if [ "$hours" -gt 0 ]; then
        duration_str="${hours}h${mins}m"
    elif [ "$mins" -gt 0 ]; then
        duration_str="${mins}m"
    fi
fi

# ===== Burn rate =====
# Per-session file: two concurrent sessions sharing one history produced nonsense velocity.
burn_file="/tmp/claude/burn-${session_id}"
burn_str=""

echo "${now_epoch},${current}" >> "$burn_file" 2>/dev/null

if [ -f "$burn_file" ]; then
    cutoff=$(( now_epoch - 120 ))
    awk -F, -v cutoff="$cutoff" '$1 >= cutoff' "$burn_file" > "${burn_file}.tmp" 2>/dev/null
    mv "${burn_file}.tmp" "$burn_file" 2>/dev/null

    first_line=$(head -1 "$burn_file" 2>/dev/null)
    last_line=$(tail -1 "$burn_file" 2>/dev/null)
    if [ -n "$first_line" ] && [ -n "$last_line" ] && [ "$first_line" != "$last_line" ]; then
        first_epoch=${first_line%%,*}; first_tokens=${first_line##*,}
        last_epoch=${last_line%%,*};   last_tokens=${last_line##*,}
        elapsed=$(( last_epoch - first_epoch ))
        token_delta=$(( last_tokens - first_tokens ))
        if [ "$elapsed" -gt 5 ] && [ "$token_delta" -gt 0 ]; then
            tpm=$(awk "BEGIN {v=$token_delta/$elapsed*60;
                if(v>=1000000) printf \"%.1fm\",v/1000000;
                else if(v>=1000) printf \"%.1fk\",v/1000;
                else printf \"%.0f\",v}")
            burn_str="${dim}↑${tpm}/m${reset}"
        fi
    fi
fi

# ===== Segment A: modes · model · git · context · burn =====
seg_a=""

if [ -n "$vim_mode" ]; then
    case "$vim_mode" in
        NORMAL) seg_a+="${dim}[${reset}${green}NOR${reset}${dim}]${reset} " ;;
        INSERT) seg_a+="${dim}[${reset}${orange}INS${reset}${dim}]${reset} " ;;
        *)      seg_a+="${dim}[${reset}${vim_mode}${dim}]${reset} " ;;
    esac
fi
[ -n "$worktree" ]   && seg_a+="${dim}[${reset}${magenta}wt:${worktree}${reset}${dim}]${reset} "
[ -n "$agent_name" ] && seg_a+="${dim}[${reset}${cyan}agent:${agent_name}${reset}${dim}]${reset} "
[ -n "$style_name" ] && seg_a+="${dim}[${reset}${blue}${style_name}${reset}${dim}]${reset} "
[ -n "$no_think" ]   && seg_a+="${dim}[${reset}${orange}nothink${reset}${dim}]${reset} "

# Model chip. Fast mode is a word, not an emoji — ⚡ is double-width and overlaps.
if [ -n "$model_name" ]; then
    seg_a+="${magenta}◆${model_name}${reset}"
    [ -n "$effort_level" ] && seg_a+=" ${dim}${effort_level}${reset}"
    [ -n "$fast_mode" ] && seg_a+=" ${yellow}fast${reset}"
    seg_a+="${gsep}"
fi

# One git call gives branch, ahead/behind and dirty state.
if [ -n "$cwd" ]; then
    display_dir="${cwd##*/}"
    seg_a+="${cyan}${display_dir}${reset}"

    git_out=$(git -C "${cwd}" status --porcelain=v2 --branch -uno 2>/dev/null)
    if [ -n "$git_out" ]; then
        IFS='|' read -r git_branch git_ahead git_behind git_dirty <<< "$(awk '
            $1=="#" && $2=="branch.head" {b=$3}
            $1=="#" && $2=="branch.ab"   {a=substr($3,2); bh=substr($4,2)}
            $1!="#"                      {d=1}
            END {printf "%s|%d|%d|%d", b, a, bh, d}
        ' <<< "$git_out")"

        if [ -n "$git_branch" ]; then
            seg_a+="${dim}@${reset}${green}${git_branch}${reset}"
            # "*" is the universal dirty marker; "●" collided with the context dots.
            [ "$git_dirty" = "1" ] && seg_a+="${orange}*${reset}"
            [ "$git_ahead" -gt 0 ] 2>/dev/null  && seg_a+=" ${blue}↑${git_ahead}${reset}"
            [ "$git_behind" -gt 0 ] 2>/dev/null && seg_a+=" ${orange}↓${git_behind}${reset}"
        fi
    fi
fi

ctx_bar=$(progress_dots "$pct_used")
pct_color=$(usage_color "$pct_used")
seg_a+="${gsep}${ctx_bar} ${pct_color}${pct_used}%${reset} ${dim}${used_tokens}/${total_tokens}${reset}"
[ -n "$burn_str" ] && seg_a+=" ${burn_str}"

# ===== Segment B: $cost · duration =====
seg_b=""
[ -n "$cost_str" ] && seg_b+="${green}${cost_str}${reset}"
if [ -n "$duration_str" ]; then
    [ -n "$seg_b" ] && seg_b+="${sep}"
    seg_b+="${dim}${duration_str}${reset}"
fi

# ===== Segment C: rate limits =====
# 5h/7d now arrive in the statusline payload — no API call, no cache, never stale.
seg_c=""
if [ "$h5_pct" -ge 0 ] 2>/dev/null; then
    h5_color=$(usage_color "$h5_pct")
    seg_c+="${white}5h${reset} ${h5_color}${h5_pct}%${reset}"
    h5_eta=$(fmt_eta "$h5_reset" "$now_epoch")
    # Plain parens, not a glyph: ⟳ (U+27F3) is ambiguous-width and overlapped the next char.
    [ -n "$h5_eta" ] && seg_c+=" ${dim}(${h5_eta})${reset}"
fi
if [ "$d7_pct" -ge 0 ] 2>/dev/null; then
    [ -n "$seg_c" ] && seg_c+="${sep}"
    d7_color=$(usage_color "$d7_pct")
    seg_c+="${white}7d${reset} ${d7_color}${d7_pct}%${reset}"
fi
[ -z "$seg_c" ] && seg_c="${dim}limits n/a${reset}"

# ===== Extra credits =====
# Only fetched when a window is nearly spent — that is the only time extra credits matter,
# so a normal session never pays for the keychain read or the network round trip.
if { [ "$h5_pct" -ge 80 ] 2>/dev/null || [ "$d7_pct" -ge 85 ] 2>/dev/null; }; then
    get_oauth_token() {
        if [ -n "$CLAUDE_CODE_OAUTH_TOKEN" ]; then echo "$CLAUDE_CODE_OAUTH_TOKEN"; return 0; fi
        local blob token
        if command -v security >/dev/null 2>&1; then
            blob=$(security find-generic-password -s "Claude Code-credentials" -w 2>/dev/null)
            token=$(printf '%s' "$blob" | jq -r '.claudeAiOauth.accessToken // empty' 2>/dev/null)
            [ -n "$token" ] && { echo "$token"; return 0; }
        fi
        if [ -f "${HOME}/.claude/.credentials.json" ]; then
            token=$(jq -r '.claudeAiOauth.accessToken // empty' "${HOME}/.claude/.credentials.json" 2>/dev/null)
            [ -n "$token" ] && { echo "$token"; return 0; }
        fi
        echo ""
    }

    cache_file="/tmp/claude/statusline-usage-cache.json"
    usage_data=""
    if [ -f "$cache_file" ]; then
        cache_mtime=$(stat -f %m "$cache_file" 2>/dev/null || stat -c %Y "$cache_file" 2>/dev/null)
        [ $(( now_epoch - cache_mtime )) -lt 300 ] && usage_data=$(cat "$cache_file" 2>/dev/null)
    fi
    if [ -z "$usage_data" ]; then
        token=$(get_oauth_token)
        if [ -n "$token" ]; then
            response=$(curl -s --max-time 5 \
                -H "Accept: application/json" \
                -H "Authorization: Bearer $token" \
                -H "anthropic-beta: oauth-2025-04-20" \
                -H "User-Agent: claude-code/2.1.34" \
                "https://api.anthropic.com/api/oauth/usage" 2>/dev/null)
            if [ -n "$response" ] && printf '%s' "$response" | jq -e . >/dev/null 2>&1; then
                usage_data="$response"
                echo "$response" > "$cache_file"
            fi
        fi
    fi

    if [ -n "$usage_data" ] && [ "$(printf '%s' "$usage_data" | jq -r '.extra_usage.is_enabled // false')" = "true" ]; then
        extra_pct=$(printf '%s' "$usage_data" | jq -r '.extra_usage.utilization // 0' | awk '{printf "%.0f", $1}')
        extra_used=$(printf '%s' "$usage_data" | jq -r '.extra_usage.used_credits // 0' | LC_NUMERIC=C awk '{printf "%.2f", $1/100}')
        extra_limit=$(printf '%s' "$usage_data" | jq -r '.extra_usage.monthly_limit // 0' | LC_NUMERIC=C awk '{printf "%.2f", $1/100}')
        extra_color=$(usage_color "$extra_pct")
        seg_c+="${sep}${white}extra${reset} ${extra_color}\$${extra_used}/\$${extra_limit}${reset}"
    fi
fi

# ===== Output =====
out="${seg_a}"
[ -n "$seg_b" ] && out+="${gsep}${seg_b}"
out+="${gsep}${seg_c}"
printf "%b" "$out"

exit 0

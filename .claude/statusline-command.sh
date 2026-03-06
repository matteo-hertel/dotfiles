#!/bin/bash
# Two-line statusline:
# Line 1: Clawd spinner | dir@branch (changes) | context bar + tokens | burn rate → ETA | effort | [vim/wt/agent]
# Line 2: $cost · duration · +lines -lines | 5h % | 7d % | extra

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
    local empty=$(( 5 - filled ))
    local bar_color
    bar_color=$(usage_color "$pct")
    local bar=""
    local i
    for (( i=0; i<filled; i++ )); do bar+="●"; done
    for (( i=0; i<empty; i++ )); do bar+="○"; done
    echo "${bar_color}${bar}${reset}"
}

sep=" ${dim}·${reset} "
gsep=" ${dim}│${reset} "
mkdir -p /tmp/claude

# ===== Extract JSON data =====

# Context window
size=$(echo "$input" | jq -r '.context_window.context_window_size // 200000')
[ "$size" -eq 0 ] 2>/dev/null && size=200000

input_tokens=$(echo "$input" | jq -r '.context_window.current_usage.input_tokens // 0')
cache_create=$(echo "$input" | jq -r '.context_window.current_usage.cache_creation_input_tokens // 0')
cache_read=$(echo "$input" | jq -r '.context_window.current_usage.cache_read_input_tokens // 0')
current=$(( input_tokens + cache_create + cache_read ))

used_tokens=$(format_tokens $current)
total_tokens=$(format_tokens $size)

if [ "$size" -gt 0 ]; then
    pct_used=$(( current * 100 / size ))
else
    pct_used=0
fi

# Session cost
cost_usd=$(echo "$input" | jq -r '.cost.total_cost_usd // 0')
cost_str=""
if [ -n "$cost_usd" ] && [ "$cost_usd" != "0" ] && [ "$cost_usd" != "null" ]; then
    cost_str=$(printf '$%.2f' "$cost_usd")
fi

# Session duration
duration_ms=$(echo "$input" | jq -r '.cost.total_duration_ms // 0')
duration_str=""
duration_s=0
if [ -n "$duration_ms" ] && [ "$duration_ms" != "0" ] && [ "$duration_ms" != "null" ]; then
    duration_s=$(( duration_ms / 1000 ))
    hours=$(( duration_s / 3600 ))
    mins=$(( (duration_s % 3600) / 60 ))
    if [ "$hours" -gt 0 ]; then
        duration_str="${hours}h${mins}m"
    elif [ "$mins" -gt 0 ]; then
        duration_str="${mins}m"
    fi
fi

# Lines changed
lines_added=$(echo "$input" | jq -r '.cost.total_lines_added // 0')
lines_removed=$(echo "$input" | jq -r '.cost.total_lines_removed // 0')

# Vim mode
vim_mode=$(echo "$input" | jq -r '.vim.mode // empty')

# Agent name
agent_name=$(echo "$input" | jq -r '.agent.name // empty')

# Worktree
worktree_name=$(echo "$input" | jq -r '.worktree.name // empty')

# ===== Burn rate + ETA =====
# Track token snapshots to compute velocity
burn_file="/tmp/claude/burn-history"
now_epoch=$(date +%s)
burn_str=""

# Append current snapshot: epoch,tokens
echo "${now_epoch},${current}" >> "$burn_file" 2>/dev/null

# Keep only last 60 seconds of data (trim old entries)
if [ -f "$burn_file" ]; then
    cutoff=$(( now_epoch - 120 ))
    awk -F, -v cutoff="$cutoff" '$1 >= cutoff' "$burn_file" > "${burn_file}.tmp" 2>/dev/null
    mv "${burn_file}.tmp" "$burn_file" 2>/dev/null

    # Calculate velocity from oldest to newest entry in window
    first_line=$(head -1 "$burn_file" 2>/dev/null)
    last_line=$(tail -1 "$burn_file" 2>/dev/null)
    if [ -n "$first_line" ] && [ -n "$last_line" ] && [ "$first_line" != "$last_line" ]; then
        first_epoch=$(echo "$first_line" | cut -d, -f1)
        first_tokens=$(echo "$first_line" | cut -d, -f2)
        last_epoch=$(echo "$last_line" | cut -d, -f1)
        last_tokens=$(echo "$last_line" | cut -d, -f2)
        elapsed=$(( last_epoch - first_epoch ))
        token_delta=$(( last_tokens - first_tokens ))
        if [ "$elapsed" -gt 5 ] && [ "$token_delta" -gt 0 ]; then
            tpm_display=$(awk "BEGIN {v=$token_delta/$elapsed*60; if(v>=1000) printf \"%.1fk\",v/1000; else printf \"%.0f\",v}")
            burn_str="${dim}${tpm_display}/min${reset}"
        fi
    fi
fi

# ===== Build segments =====

# Segment A: indicators + dir/git + context + burn rate
seg_a=""

if [ -n "$vim_mode" ]; then
    case "$vim_mode" in
        NORMAL) seg_a+=" ${dim}[${reset}${green}NOR${reset}${dim}]${reset}" ;;
        INSERT) seg_a+=" ${dim}[${reset}${orange}INS${reset}${dim}]${reset}" ;;
        *)      seg_a+=" ${dim}[${reset}${vim_mode}${dim}]${reset}" ;;
    esac
fi
if [ -n "$worktree_name" ]; then
    seg_a+=" ${dim}[${reset}${magenta}wt:${worktree_name}${reset}${dim}]${reset}"
fi
if [ -n "$agent_name" ]; then
    seg_a+=" ${dim}[${reset}${cyan}agent:${agent_name}${reset}${dim}]${reset}"
fi

cwd=$(echo "$input" | jq -r '.cwd // empty')
if [ -n "$cwd" ]; then
    display_dir="${cwd##*/}"
    git_branch=$(git -C "${cwd}" rev-parse --abbrev-ref HEAD 2>/dev/null)
    [ -n "$seg_a" ] && seg_a+="${sep}"
    seg_a+="${cyan}${display_dir}${reset}"
    if [ -n "$git_branch" ]; then
        seg_a+="${dim}@${reset}${green}${git_branch}${reset}"
        git_stat=$(git -C "${cwd}" diff --numstat 2>/dev/null | awk '{a+=$1; d+=$2} END {if (a+d>0) printf "+%d -%d", a, d}')
        if [ -n "$git_stat" ]; then
            seg_a+=" ${dim}(${reset}${green}${git_stat%% *}${reset} ${red}${git_stat##* }${reset}${dim})${reset}"
        fi
    fi
fi

ctx_bar=$(progress_dots "$pct_used")
pct_color=$(usage_color "$pct_used")
seg_a+="${gsep}${ctx_bar} ${pct_color}${pct_used}%${reset} ${dim}${used_tokens}/${total_tokens}${reset}"

if [ -n "$burn_str" ]; then
    seg_a+=" ${burn_str}"
fi

# Segment B: session stats (cost · duration · lines)
seg_b=""
need_sep=""

if [ -n "$cost_str" ]; then
    seg_b+="${green}${cost_str}${reset}"
    need_sep=1
fi

if [ -n "$duration_str" ]; then
    [ -n "$need_sep" ] && seg_b+="${sep}"
    seg_b+="${dim}${duration_str}${reset}"
    need_sep=1
fi

if [ "$lines_added" -gt 0 ] || [ "$lines_removed" -gt 0 ]; then
    [ -n "$need_sep" ] && seg_b+="${sep}"
    [ "$lines_added" -gt 0 ] && seg_b+="${green}+${lines_added}${reset}"
    if [ "$lines_removed" -gt 0 ]; then
        [ "$lines_added" -gt 0 ] && seg_b+=" "
        seg_b+="${red}-${lines_removed}${reset}"
    fi
    need_sep=1
fi

# ===== OAuth token resolution =====
get_oauth_token() {
    if [ -n "$CLAUDE_CODE_OAUTH_TOKEN" ]; then
        echo "$CLAUDE_CODE_OAUTH_TOKEN"
        return 0
    fi
    if command -v security >/dev/null 2>&1; then
        local blob
        blob=$(security find-generic-password -s "Claude Code-credentials" -w 2>/dev/null)
        if [ -n "$blob" ]; then
            local token
            token=$(echo "$blob" | jq -r '.claudeAiOauth.accessToken // empty' 2>/dev/null)
            if [ -n "$token" ] && [ "$token" != "null" ]; then
                echo "$token"
                return 0
            fi
        fi
    fi
    local creds_file="${HOME}/.claude/.credentials.json"
    if [ -f "$creds_file" ]; then
        local token
        token=$(jq -r '.claudeAiOauth.accessToken // empty' "$creds_file" 2>/dev/null)
        if [ -n "$token" ] && [ "$token" != "null" ]; then
            echo "$token"
            return 0
        fi
    fi
    if command -v secret-tool >/dev/null 2>&1; then
        local blob
        blob=$(timeout 2 secret-tool lookup service "Claude Code-credentials" 2>/dev/null)
        if [ -n "$blob" ]; then
            local token
            token=$(echo "$blob" | jq -r '.claudeAiOauth.accessToken // empty' 2>/dev/null)
            if [ -n "$token" ] && [ "$token" != "null" ]; then
                echo "$token"
                return 0
            fi
        fi
    fi
    echo ""
}

# ===== Usage limits (cached) =====
cache_file="/tmp/claude/statusline-usage-cache.json"
cache_max_age=60
mkdir -p /tmp/claude

needs_refresh=true
usage_data=""

if [ -f "$cache_file" ]; then
    cache_mtime=$(stat -c %Y "$cache_file" 2>/dev/null || stat -f %m "$cache_file" 2>/dev/null)
    now=$(date +%s)
    cache_age=$(( now - cache_mtime ))
    if [ "$cache_age" -lt "$cache_max_age" ]; then
        needs_refresh=false
        usage_data=$(cat "$cache_file" 2>/dev/null)
    fi
fi

if $needs_refresh; then
    token=$(get_oauth_token)
    if [ -n "$token" ] && [ "$token" != "null" ]; then
        response=$(curl -s --max-time 10 \
            -H "Accept: application/json" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $token" \
            -H "anthropic-beta: oauth-2025-04-20" \
            -H "User-Agent: claude-code/2.1.34" \
            "https://api.anthropic.com/api/oauth/usage" 2>/dev/null)
        if [ -n "$response" ] && echo "$response" | jq . >/dev/null 2>&1; then
            usage_data="$response"
            echo "$response" > "$cache_file"
        fi
    fi
    if [ -z "$usage_data" ] && [ -f "$cache_file" ]; then
        usage_data=$(cat "$cache_file" 2>/dev/null)
    fi
fi

# Segment C: rate limits (5h · 7d · extra)
seg_c=""
if [ -n "$usage_data" ] && echo "$usage_data" | jq -e . >/dev/null 2>&1; then
    five_hour_pct=$(echo "$usage_data" | jq -r '.five_hour.utilization // 0' | awk '{printf "%.0f", $1}')
    five_hour_color=$(usage_color "$five_hour_pct")
    seg_c+="${white}5h${reset} ${five_hour_color}${five_hour_pct}%${reset}"

    seven_day_pct=$(echo "$usage_data" | jq -r '.seven_day.utilization // 0' | awk '{printf "%.0f", $1}')
    seven_day_color=$(usage_color "$seven_day_pct")
    seg_c+="${sep}${white}7d${reset} ${seven_day_color}${seven_day_pct}%${reset}"

    extra_enabled=$(echo "$usage_data" | jq -r '.extra_usage.is_enabled // false')
    if [ "$extra_enabled" = "true" ]; then
        extra_pct=$(echo "$usage_data" | jq -r '.extra_usage.utilization // 0' | awk '{printf "%.0f", $1}')
        extra_used=$(echo "$usage_data" | jq -r '.extra_usage.used_credits // 0' | LC_NUMERIC=C awk '{printf "%.2f", $1/100}')
        extra_limit=$(echo "$usage_data" | jq -r '.extra_usage.monthly_limit // 0' | LC_NUMERIC=C awk '{printf "%.2f", $1/100}')
        if [ -n "$extra_used" ] && [ -n "$extra_limit" ] && [[ "$extra_used" != *'$'* ]] && [[ "$extra_limit" != *'$'* ]]; then
            extra_color=$(usage_color "$extra_pct")
            seg_c+="${sep}${white}extra${reset} ${extra_color}\$${extra_used}/\$${extra_limit}${reset}"
        else
            seg_c+="${sep}${white}extra${reset} ${green}enabled${reset}"
        fi
    fi
fi

# ===== Output =====
out="${seg_a}"
[ -n "$seg_b" ] && out+="${gsep}${seg_b}"
[ -n "$seg_c" ] && out+="${gsep}${seg_c}"
printf "%b" "$out"

exit 0

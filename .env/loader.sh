#!/bin/bash
CURRENT_DIR=$PWD
cd "$(dirname "$0")"
for f in "$PWD"/aliases/*; do
   source "$f"
done
for f in "$PWD"/tmux/*; do
   source "$f"
done

cd "$CURRENT_DIR" > /dev/null

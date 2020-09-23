#!/bin/bash

if ! command -v wrk2 >/dev/null 2>&1; then
  if [ -f "/etc/arch-release" ]; then
    yay -S wrk2-git
  fi
fi

wrk2 -t5 -c200 -d30s -R2000 -s scripts/payload.lua http://localhost:8000/

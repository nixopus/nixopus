#!/bin/bash
cd /Users/shaale/nixopus/api
/Users/shaale/.gvm/gos/go1.25/bin/go build -o bin/nixopus-api . 2>/tmp/go_build_stderr.log
echo "Exit: $?"
cat /tmp/go_build_stderr.log

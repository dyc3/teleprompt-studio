#!/bin/bash

# When the program crashes, it usually doesn't write the ANSI escape sequences to ignore mouse events again.
# If that happens, you can run this script to fix it.

echo -e "\x1b[?1006l\x1b[?1015l\x1b[?1002l\x1b[?1000l\x1b[?25h"
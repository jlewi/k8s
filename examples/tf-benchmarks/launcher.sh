#!/bin/bash
set -x
echo Running launcher.sh
python launcher.py $@
chmod a+x /opt/run_benchmarks.sh
/opt/run_benchmarks.sh
#!/bin/bash

set -x

# Create the launch script
python create_launch_command.py $@ > /opt/run_benchmarks.sh
chmod a+x /opt/run_benchmarks.sh
/opt/run_benchmarks.sh
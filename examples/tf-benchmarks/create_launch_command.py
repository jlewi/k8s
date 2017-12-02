#!/usr/bin/python
"""A launcher suitable for invoking tf_cnn_benchmarks using TfJob.

All the launcher does is turn TF_CONFIG environment variable
into extra arguments to append to the command line.
"""

import json
import os
import subprocess
import sys
import time

if __name__ == "__main__":
  tf_config = os.environ.get('TF_CONFIG')
  tf_config_json = json.loads(tf_config)
  cluster = tf_config_json.get('cluster')
  job_name = tf_config_json.get('task', {}).get('type')
  task_index = tf_config_json.get('task', {}).get('index')

  command = sys.argv[1:]
  ps_hosts = ",".join(cluster.get("ps", []))
  worker_hosts = ",".join(cluster.get("worker", []))
  command.append("--job_name=" + job_name)
  command.append("--ps_hosts=" + ps_hosts)
  command.append("--worker_hosts=" + worker_hosts)
  command.append("--task_index={0}".format(task_index))

  if job_name.lower() == "master":
    print("#!/bin/bash")
    print("echo master runs for ever.")
    print("tail -f /dev/null")

  print("#!/bin/bash")
  print(" ".join(command))

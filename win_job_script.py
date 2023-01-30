# -*- coding: utf-8 -*-
import sys
import os
import subprocess
import json
import platform
import socket
import logging
import time

from google.cloud import bigquery
from google.cloud import storage


# init logger
logging.basicConfig(
    filename='C:/python_script_log.log',
    format="%(asctime)s,%(msecs)d %(name)s %(levelname)s %(message)s",
    datefmt='%H:%M:%S'
)

logging.info("test.py started")

# define job status
wip = "Work in Progress"
done = "Completed"

logging.info("read configration")
# get configurations
fd_conf = open('win_job_conf.txt', 'r')
project_id = fd_conf.readline().strip()
jobname = fd_conf.readline().strip()
cmd_exe = fd_conf.readline().strip()
exe_log_path = fd_conf.readline().strip()
fd_bucketname = open('win_bucket_conf.txt', 'r')
bucketname = fd_bucketname.readline().strip()

dataset = "gametest"
table = "job-result"
table_id = project_id + "." + dataset + "." + table

hostname = socket.gethostname()
os_ver = platform.platform()

logging.info("Job started & update BQ")
# Job started & update BQ
insert = f"""
    INSERT INTO
    `{table_id}` (job_name,
        hostname, os_version, status)
    VALUES
    ('{jobname}', '{hostname}', '{os_ver}', '{wip}')
""".format(table_id, jobname, hostname, os_ver, wip)

bq = bigquery.Client()

bq.query(insert)

logging.info("BQ updated")

# patch
logging.info("Application Set Up")
# t = subprocess.run(["py", "exe.py"], capture_output=True)
t = subprocess.run(cmd_exe, capture_output=True)
result_code = t.returncode

if result_code == 0:
    result = "success"
else:
    result = "fail"

message = jobname + "/" + hostname + "." + str(time.time()) + ".log" 

logging.info("Updating BQ")
# update job status in BQ

update = f"""
    UPDATE
    `{table_id}`
    SET status='{done}', result='{result}', message="{message}"
    where
    job_name='{jobname}' AND hostname='{hostname}'
""".format(table_id, done, result, message, jobname, hostname)

update = f"""UPDATE
  `{table_id}`
SET
  status='{done}',
  result='{result}',
  message='{message}'
WHERE
  job_name='{jobname}'
  AND hostname='{hostname}'""".format(table_id, done, result, message, jobname, hostname)

bq.query(update)

# upload log file to GCS
gcs = storage.Client()
bucket = gcs.bucket(bucketname)
blob = bucket.blob(jobname + "/" + hostname + "." + str(time.time()) + ".log")
blob.upload_from_filename(exe_log_path)

logging.info("qtest.py done")

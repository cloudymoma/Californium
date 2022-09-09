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

os.environ['GOOGLE_APPLICATION_CREDENTIALS'] = "./windev.json"

# init logger
logging.basicConfig(
    level=logging.DEBUG,
    format="%(asctime)s,%(msecs)d %(name)s %(levelname)s %(message)s",
    handlers=[
        logging.FileHandler("testpy_debug.log"),
        logging.StreamHandler(sys.stdout)
    ]
)

logging.info("test.py started")

# define job status
wip = "Work in Progress"
done = "Completed"

# get configurations
fd_conf = open('./conf', 'r')
project_id = fd_conf.readline().strip()
jobname = fd_conf.readline().strip()
cmd_exe = fd_conf.readline.strip()
exe_log_path = fd_conf.readline.strip()

dataset = "windev"
table = "patchdata"
table_id = project_id + "." + dataset + "." + table

hostname = socket.gethostname()
os_ver = platform.platform()

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

# patch
# t = subprocess.run(["py", "exe.py"], capture_output=True)
t = subprocess.run(cmd_exe, capture_output=True)
out = t.stdout.decode("utf-8")
err = t.stderr.decode("utf-8")

message = f"""stdout: '{out}' stderr: '{err}'""".format(out, err).replace('\r\n', ' ').replace('\n', ' ')

result = str(t.returncode)

logging.debug("====== message ======")
logging.debug(message)
logging.debug("====== message end ======")

# update job status in BQ
update = f"""
    UPDATE
    `{table_id}`
    SET status='{done}', result='{result}', message="{message}"
    where
    job_name='{jobname}' AND hostname='{hostname}'
""".format(table_id, done, result, message, jobname, hostname)

bq.query(update)

# upload log file to GCS
gcs = storage.Client()
fd_bucketname = open('./bucketname', 'r')
bucketname = fd_bucketname.readline().strip()
bucket = gcs.bucket(bucketname)
blob = bucket.blob(jobname + "/" + hostname + "." + str(time.time()) + ".log")
blob.upload_from_filename(exe_log_path)

logging.info("test.py done")

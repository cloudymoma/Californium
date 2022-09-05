# -*- coding: utf-8 -*-

import sys
import os
import subprocess
import json
import platform
import socket

from google.cloud import bigquery

os.environ['GOOGLE_APPLICATION_CREDENTIALS'] = "./windev.json"

print("test.py started")

# define job status
wip = "Work in Progress"
done = "Completed"

# get configurations
fd_conf = open('./conf', 'r')
project_id = fd_conf.readline().strip()
jobname = fd_conf.readline().strip()
dataset = "windev"
table = "patchdata"
table_id = project_id + "." + dataset + "." + table

hostname = socket.gethostname()
os_ver = platform.platform()

insert = f"""
    INSERT INTO
    `{table_id}` (job_name, 
        hostname, os_version, status)
    VALUES
    ('{jobname}', '{hostname}', '{os_ver}', '{wip}')
""".format(table_id, jobname, hostname, os_ver, wip)

bq = bigquery.Client()

bq.query(insert)

# du-hast-mich.windev.patchdata

"""
    #"job_name": os.getenv('JOB_NAME'),
    #"instance_name": os.getenv('INST_NAME'),
d = {
    "job_name": jobname,
    "hostname": socket.gethostname(),
    "os_version": platform.platform(),
    "status": wip,
    "result": result,
    "message": message
}

print("json data: %s" % (json.dumps(d)))

rows_to_insert = [
    d,
]

# app_test.job_result
r = bq.insert_rows_json(os.getenv('TABLE_ID', rows_to_insert))
if r == []:
    print("job done")
else:
    print("errors: {}".format(r))
"""

# patch
t = subprocess.run(["py", "exe.py"], capture_output=True)
out = t.stdout.decode("utf-8")
err = t.stderr.decode("utf-8")

message = f"""stdout: '{out}' stderr: '{err}'""".format(out, err).replace('\r\n', ' ').replace('\n', ' ')

result = str(t.returncode)

update = f"""
    UPDATE
    `{table_id}`
    SET status='{done}', result='{result}', message="{message}"
    where
    job_name='{jobname}' AND hostname='{hostname}'
""".format(table_id, done, result, message, jobname, hostname)

bq.query(update)

print("test.py done")

REM set /p jobname=<jobname
set /p bucketname=<bucketname

gsutil cp gs://%bucketname%/python-3.10.6-amd64.exe .
REM gsutil cp gs://%bucketname%/test.py .

REM echo "Installing Python 3.10.6"
REM python-3.10.6-amd64.exe /quiet InstallAllUsers=1 PrependPath=1 Include_test=0

REM echo "Installing bigquery client for python"
REM py -m pip install -U google-cloud-bigquery

echo "Patching"
py test.py

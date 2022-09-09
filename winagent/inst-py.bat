set /p bucketname=<bucketname

call gsutil -m cp gs://%bucketname%/python-3.10.6-amd64.exe gs://%bucketname%/test.py gs://%bucketname%/windev.json gs://%bucketname%/exe.py .

echo "Installing Python 3.10.6"
python-3.10.6-amd64.exe /quiet InstallAllUsers=1 PrependPath=1 Include_test=0

echo "Installing bigquery client for python"
py -m pip install -U google-cloud-bigquery google-cloud-storage

echo "Patching"
py test.py

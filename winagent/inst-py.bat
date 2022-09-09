set /p bucketname=<C:\bucketname

echo "Downloading files from gcs" > C:\inst-batlog
call gsutil -m cp gs://%bucketname%/python-3.10.6-amd64.exe gs://%bucketname%/conf gs://%bucketname%/test.py .

echo "Installing Python 3.10.6" >> C:\inst-batlog
python-3.10.6-amd64.exe /quiet InstallAllUsers=1 PrependPath=1 Include_test=0
echo "Successfully installed python" >> C:\inst-batlog

echo "Installing  google-cloud-bigquery and google-cloud-storage " >> C:\inst-batlog
py -m pip install -U google-cloud-bigquery google-cloud-storage
echo "Successully installed  google-cloud-bigquery and google-cloud-storage " >> C:\inst-batlog

echo "Patching" >> C:\inst-batlog
py test.py
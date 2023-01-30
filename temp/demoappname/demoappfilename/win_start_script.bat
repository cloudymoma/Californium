echo "Downloading files from gcs" > C:\inst-batlog
call gsutil -m cp -r gs://mongodb-on-gke-democli-xvlb/demoappname/demoappfilename C:\
cd C:\demoappfilename
echo "Installing Python 3.10.7" >> C:\inst-batlog
python.exe /quiet InstallAllUsers=1 PrependPath=1 Include_test=0
echo "Successfully installed python" >> C:\inst-batlog
echo "Installing  google-cloud-bigquery and google-cloud-storage " >> C:\inst-batlog
py -m pip install -U google-cloud-bigquery google-cloud-storage
echo "Successully installed  google-cloud-bigquery and google-cloud-storage " >> C:\inst-batlog
echo "Patching" >> C:\inst-batlog
py win_job_script.py

import os
import shutil
import urllib.request
from google.cloud import storage


def py_installer_download(temp_dir, app_name, app_file_name):
    isExist = os.path.exists(f"{temp_dir}/{app_name}/{app_file_name}")
    if not isExist:
        os.makedirs(f"{temp_dir}/{app_name}/{app_file_name}")
        print("Create temp dir")
    else:
        print("temp dir already exist")
    print("downloading python.exe file")
    url_link = "https://www.python.org/ftp/python/3.10.7/python-3.10.7-amd64.exe"
    urllib.request.urlretrieve(url_link, f"{temp_dir}/{app_name}/{app_file_name}/python.exe")
    print("finish python.exe download")


def win_start_script(temp_dir, bucket_name, app_name, app_file_name):
    print("saving windows start up script")
    win_start_script = """echo "Downloading files from gcs" > C:\inst-batlog
call gsutil -m cp -r gs://{bucket_name}/{app_name}/{app_file_name} C:\\
cd C:\{app_name}
echo "Installing Python 3.10.7" >> C:\inst-batlog
python.exe /quiet InstallAllUsers=1 PrependPath=1 Include_test=0
echo "Successfully installed python" >> C:\inst-batlog
echo "Installing  google-cloud-bigquery and google-cloud-storage " >> C:\inst-batlog
py -m pip install -U google-cloud-bigquery google-cloud-storage
echo "Successully installed  google-cloud-bigquery and google-cloud-storage " >> C:\inst-batlog
echo "Patching" >> C:\inst-batlog
py win_job_script.py
""".format(bucket_name=bucket_name, app_name=app_name, app_file_name=app_file_name)
    with open(f"{temp_dir}/{app_name}/{app_file_name}/win_start_script.bat", "w") as f:
      f.write(win_start_script)



def win_bucket_conf(temp_dir, app_name, bucket_name, app_file_name):
  print("saving bucket configuration file")
  win_bucke_conf = f"{bucket_name}"
  with open(f"{temp_dir}/{app_name}/{app_file_name}/win_bucket_conf.txt", "w") as f:
    f.write(win_bucke_conf)


def win_job_conf(project_id, job_name, cmd, temp_dir, app_name):
    print("saving bucket configuration file")
    win_job_conf = f"""{project_id}
{job_name}
{cmd}
C:\patch.log"""
    print(win_job_conf)
    with open(f"{temp_dir}/{app_name}/{app_file_name}/win_job_conf.txt", "w") as f:
      f.write(win_job_conf)


def win_job_script(temp_dir, app_name, app_file_name):
    print("saving win job python script")
    src_file = os.path.expanduser('~') + "/win_job_script.py"
    dest_file = f"{temp_dir}/{app_name}/{app_file_name}/win_job_script.py"
    shutil.copy(src_file, dest_file)


def upload_blob(bucket_name, temp_dir, app_name, app_file_name):
    storage_client = storage.Client()
    bucket = storage_client.get_bucket(bucket_name)
    files = os.listdir(temp_dir + '/'+app_name + '/' + app_file_name)
    for file in files:
        source_file_name = f"{temp_dir}/{app_name}/{app_file_name}/{file}"
        destination_blob_name = f"{app_name}/{app_file_name}/{file}"
        blob = bucket.blob(destination_blob_name)
        blob.upload_from_filename(source_file_name)
        print(
          f"File {source_file_name} uploaded to {destination_blob_name}."
        )


def all_files_upload(project_id, job_name, temp_dir, bucket_name, app_name, app_file_name, cmd):
    py_installer_download(temp_dir, app_name, app_file_name)
    win_start_script(temp_dir, bucket_name, app_name, app_file_name)
    win_bucket_conf(temp_dir, app_name, bucket_name, app_file_name)
    win_job_conf(project_id, job_name, cmd, temp_dir, app_name, app_file_name)
    win_job_script(temp_dir, app_name, app_file_name)
    upload_blob(bucket_name, temp_dir, app_name, app_file_name)

# Californium Command Line Tool User Guide

Californium is an open source game package quality test framework based on GCP(google cloud platform) service. It uses C/S achitecture, which server backend side help you to handle game package test job create/delete/query, on the other hand this command line tool acts as client side, help users to create/delete GCP resources serve this test framework and interactive job management with server side by simple commands and paramters. 

* [Quick Started](#quick-started)

## Quick Started
### Prepare the Google Cloud environment
1. confirm the GCP project is created, if not refer the [project management](https://cloud.google.com/resource-manager/docs/creating-managing-projects#creating_a_project) docs to create the GCP project which is used to host the Californium backend resources.
2. create service account that will be used by the command line tool to interactive with GCP API. refer the [service account management](https://cloud.google.com/iam/docs/creating-managing-service-accounts#creating) docs. 
3. grant the project owner role to the service account created in step 2. refer the [IAM role](https://cloud.google.com/iam/docs/manage-access-service-accounts#single-role) docs.
4. create/download a service account key. refer the [service account key management](https://cloud.google.com/iam/docs/creating-managing-service-account-keys#creating) docs

### Install the command line tool
The command line tool in this repo is devloped using Golang, user can clone this repo to build the client locally. 
```
cd ./Californium/californiumcli
go build .
```
or user can download the published binary
```
add the download command
```

### Create the client config file
Edit the .cobra.json file, base on the sample config file, update the projectid to the project that create before. If user has special preference for other porperties, user can set the preferred value to other porperties, for quick start, suggest to leave the default value of other porperties. 
```
{
  "bucketname": "democli",
  "datasetid": "gametest",
  "instancename": "democli",
  "machinetype": "n1-standard-1",
  "networkname": "global/networks/default",
  "projectid": "demoproject",
  "sourceimage": "projects/debian-cloud/global/images/family/debian-11",
  "zone": "us-central1-a"
}
```

### Create the job config file
Edit the job.json file, the job_name porperty is the unique name for each job. os_list porperty is the target OS platform the game package test will be running on, supported OS type is  {"win7": "win7", "win8": "win8", "win10": "win10", "win11": "win11"}, user can just select the OS platform that intested. Below is a sample that just select win7 OS.
```
{
    "job_name": "demo",
    "os_list": {"win7": "win7"}
}
```

### Set service account key file environment variable path
Replace the /PATH/serviceaccountkey.json to the file path that service account key file located.
```
export GOOGLE_APPLICATION_CREDENTIALS='/PATH/serviceaccountkey.json'
```

### Create Californium backend resource

```
./californiumcli serve --config ".cobra.json"
```

### Create Job

```
./californiumcli job create --config ".cobra.json" --jobconfig "job.json"
```
Note: if user install the client tool outside the backend server, please set the -a (backenserver ip) and -l (backend server port) accordingly 

### Query Job

```
./californiumcli job query --config ".cobra.json" --jobconfig "job.json"
```
Note: if user install the client tool outside the backend server, please set the -a (backenserver ip) and -l (backend server port) accordingly 

### Get Job Result

```
./californiumcli job result --config ".cobra.json" --jobconfig "job.json"
```
Note: if user install the client tool outside the backend server, please set the -a (backenserver ip) and -l (backend server port) accordingly 

### Delete Job Result

```
./californiumcli job delete --config ".cobra.json" --jobconfig "job.json"
```
Note: if user install the client tool outside the backend server, please set the -a (backenserver ip) and -l (backend server port) accordingly 

### Clean Californium backend resource

```
./californiumcli clean --config ".cobra.json"
```

### Other usage
Command line tool provide help guide, user can use -h parameter to get the help guide
```
./californiumcli -h
```

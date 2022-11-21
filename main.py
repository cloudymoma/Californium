import os
import sys
import json
import time

from fastapi import FastAPI
from typing import Union
from fastapi import FastAPI
from pydantic import BaseModel
from typing import Dict

import pulumi
from pulumi import automation as auto
from pulumi import Output
from pulumi_gcp import storage
from pulumi_gcp import compute

from google.cloud import compute_v1
from google.cloud import bigquery


image_list_origin = {
    "win7": "windows-cloud/windows-server-2016-dc-v20220902",
    "win8": "windows-cloud/windows-server-2012-r2-dc-v20220902",
    "win10": "windows-cloud/windows-server-2016-dc-v20220902",
    "win11": "windows-cloud/windows-server-2016-dc-v20220902"
}

def job_status_init(job_name):
    client = bigquery.Client()
    insert = f"""
        INSERT INTO
        `game_test.job_status` (job_name,
            job_status,
            message)
        VALUES
        ('{job_name}', 'starting', 'test job is starting')
    """.format(job_name=job_name)
    client.query(insert)
    print("initialize job status")

def job_stauts_update(job_name):
    client = bigquery.Client()
    update = f"""
    UPDATE
    `game_test.job_status`
    SET
    job_status="on progress", message="test job is on progress"
    WHERE
    job_name='{job_name}'""".format(job_name=job_name)
    client.query(update)
    print("update job status")

def job_status_check(job_name):
    client = bigquery.Client()
    update = f"""
    SELECT
    *
    FROM
    `game_test.job_status`
    WHERE
    job_name = '{job_name}'""".format(job_name=job_name)
    query_results = client.query(update)
    query_records = [dict(row) for row in query_results]
    final_results = json.dumps(str(query_records))
    return final_results 

def job_result_check(job_name):
    client = bigquery.Client()
    query = f"""
    SELECT
    *
    FROM
    `game_test.job_result`
    WHERE
    job_name='{job_name}'""".format(job_name=job_name)
    query_results = client.query(query)
    query_records = [dict(row) for row in query_results]
    final_results = json.dumps(str(query_records))
    return final_results 

def job_status_delete(job_name):
    client = bigquery.Client()
    update = f"""
    UPDATE
    `game_test.job_status`
    SET
    job_status="deleted", message="test job has been deleted"
    WHERE
    job_name='{job_name}'""".format(job_name=job_name)
    client.query(update)
    print("update job status")

def instance_status_check(image_list, project_id, zone):
    instance_list = []
    for image_key in image_list:
        instance_list.append(image_key)
    client = compute_v1.InstancesClient()
    status_running = 0
    while (status_running < len(instance_list)):
        for instance in instance_list:
            instance_obj = client.get(project=project_id, zone=zone, instance=instance)
            instance_status = instance_obj.status
            if instance_status == "RUNNING":
                status_running += 1
            else:
                pass
        print("waiting for all instances become ready")
    print("All instances are in running status now")



def job_create(program, job_name, project_id, region, zone):
    project_name = job_name
    os.environ["PULUMI_CONFIG_PASSPHRASE"] = ""
    project_setting = auto.ProjectSettings(name=project_name, runtime="python", backend=auto.ProjectBackend(url="file:///pulumidemobackend"))
    stack_setting = {
    "prd": auto.StackSettings(secrets_provider="default")
    }
    localworkspace_setting = auto.LocalWorkspaceOptions(work_dir="./",secrets_provider="default" ,project_settings=project_setting, stack_settings=stack_setting)
    stack_name = job_name
    stack = auto.create_or_select_stack(stack_name=stack_name,
                                    project_name=project_name,
                                    program=program,
                                    opts=localworkspace_setting)

    print("initialize pulumi stack")
    stack.workspace.install_plugin("gcp", "v6.32.0")
    gcp_configmap_setting = {
    "gcp:project": auto.ConfigValue(value=project_id),
    "gcp:region": auto.ConfigValue(value=region),
    "gcp:zone": auto.ConfigValue(value=zone)
    }
    stack.set_all_config(gcp_configmap_setting)
    print("config pulumi stack")
    time.sleep(5)
    up_res = stack.up(on_output=print)
    print("pulumi stack has been created successfully")

def job_delete(program, job_name, project_id, region, zone):
    project_name = job_name
    os.environ["PULUMI_CONFIG_PASSPHRASE"] = ""
    project_setting = auto.ProjectSettings(name=project_name, runtime="python", backend=auto.ProjectBackend(url="file:///pulumidemobackend"))
    stack_setting = {
    "prd": auto.StackSettings(secrets_provider="default")
    }
    localworkspace_setting = auto.LocalWorkspaceOptions(work_dir="./",secrets_provider="default" ,project_settings=project_setting, stack_settings=stack_setting)
    stack_name = job_name
    stack = auto.create_or_select_stack(stack_name=stack_name,
                                    project_name=project_name,
                                    program=program,
                                    opts=localworkspace_setting)

    print("pulumi stack initialized")
    stack.workspace.install_plugin("gcp", "v6.32.0")
    gcp_configmap_setting = {
    "gcp:project": auto.ConfigValue(value=project_id),
    "gcp:region": auto.ConfigValue(value=region),
    "gcp:zone": auto.ConfigValue(value=zone)
    }
    stack.set_all_config(gcp_configmap_setting)
    up_destroy = stack.destroy(on_output=print)
    print(f'successfully delete job {job_name}'.format(job_name=job_name))
    


class os_list(BaseModel):
    win7: str = None
    win8: str = None
    win10: str = None
    win11: str = None

class job_conf(BaseModel):
    job_name: str = None
    os_list: os_list 
    project_id: str = None
    region: str = None
    zone: str = None 
    instance_type: str = None
    vpc_network: str = None
    gcs_bucket: str = None


app = FastAPI()

@app.get("/jobs/{job_name}")
async def get_handle(job_name):
    result = job_status_check(job_name=job_name)
    return result


@app.post("/jobs/")
async def post_handle(req: job_conf):
    job_name = req.job_name
    os_list = req.os_list.dict()
    project_id = req.project_id
    region = req.region
    zone = req.zone
    instance_type = req.instance_type
    vpc_network = req.vpc_network
    gcs_bucket = req.gcs_bucket
    image_list_final = {}
    for image_key in os_list:
        if os_list[image_key] is not None:
            image_list_final[image_key] = image_list_origin[image_key]


    def gcp_resource(image_list=image_list_final):
        for image_key in image_list: 
            instance_dic = {}
            instance_dic[image_key] = compute.Instance(image_key,
                name = image_key,
                machine_type=instance_type,
                zone=zone,
                boot_disk=compute.InstanceBootDiskArgs(
                    initialize_params=compute.InstanceBootDiskInitializeParamsArgs(
                        image=image_list[image_key],
                    ),
                ),
                metadata={
                    "sysprep-specialize-script-cmd": f"echo {gcs_bucket}>C:\\bucketname".format(gcs_bucket=gcs_bucket),
                    "windows-startup-script-url": f"gs://{gcs_bucket}/inst-py.bat".format(gcs_bucket=gcs_bucket),
                },
                service_account=compute.InstanceServiceAccountArgs(
                    email="baremetal-server@mongodb-on-gke.iam.gserviceaccount.com",
                    scopes=["cloud-platform"],
                ),
                network_interfaces=[compute.InstanceNetworkInterfaceArgs(
                    network=vpc_network,
                    access_configs=[compute.InstanceNetworkInterfaceAccessConfigArgs()],
                )])
            pulumi.export(f'instance-{image_key}_status'.format(image_key),  instance_dic[image_key].current_status)
    job_status_init(job_name=job_name)
    job_create(program=gcp_resource, job_name=job_name, project_id=project_id, region=region, zone=zone)
    instance_status_check(image_list=image_list_final, project_id=project_id, zone=zone)
    job_stauts_update(job_name=job_name)



@app.delete("/jobs/")
async def delete_handle(req: job_conf):
    job_name = req.job_name
    os_list = req.os_list.dict()
    project_id = req.project_id
    region = req.region
    zone = req.zone
    instance_type = req.instance_type
    vpc_network = req.vpc_network
    gcs_bucket = req.gcs_bucket
    image_list_final = {}
    for image_key in os_list:
        if os_list[image_key] is not None:
            image_list_final[image_key] = os_list[image_key]
    print(os_list)
    print(image_list_final)

    def gcp_resource(image_list=image_list_final):
        for image_key in image_list: 
            instance_dic = {}
            instance_dic[image_key] = compute.Instance(image_key,
                name = image_key,
                machine_type=instance_type,
                zone=zone,
                boot_disk=compute.InstanceBootDiskArgs(
                    initialize_params=compute.InstanceBootDiskInitializeParamsArgs(
                        image=image_list[image_key],
                    ),
                ),
                metadata={
                    "sysprep-specialize-script-cmd": f"echo {gcs_bucket}>C:\bucketname".format(gcs_bucket=gcs_bucket),
                    "windows-startup-script-url": f"gs://{gcs_bucket}/inst-py.bat".format(gcs_bucket=gcs_bucket),
                },
                service_account=compute.InstanceServiceAccountArgs(
                    email="baremetal-server@mongodb-on-gke.iam.gserviceaccount.com",
                    scopes=["cloud-platform"],
                ),
                network_interfaces=[compute.InstanceNetworkInterfaceArgs(
                    network=vpc_network,
                    access_configs=[compute.InstanceNetworkInterfaceAccessConfigArgs()],
                )])
            pulumi.export(f'instance-{image_key}_status'.format(image_key),  instance_dic[image_key].current_status)
    
    job_delete(program=gcp_resource, job_name=job_name, project_id=project_id, region=region, zone=zone)
    job_status_delete(job_name=job_name)

@app.get("/jobresult/{job_name}")
async def get_handle(job_name):
    result = job_result_check(job_name=job_name)
    return result

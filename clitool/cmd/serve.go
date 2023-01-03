/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	compute "cloud.google.com/go/compute/apiv1"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	resourcemanagerpb "google.golang.org/genproto/googleapis/cloud/resourcemanager/v3"
	"google.golang.org/protobuf/proto"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Create required backend resource(gce, gcs, bigquery)",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		projectID = viper.GetString("projectid")
		zone = viper.GetString("zone")
		instanceName = viper.GetString("instancename")
		bucketName = projectID + "-" + viper.GetString("bucketname") + "-" + RandStringBytes(4)
		datasetID = viper.GetString("datasetid")
		// tableID = viper.GetString("tableid")

		if projectID == "" || zone == "" || instanceName == "" || bucketName == "" {
			fmt.Printf("Please give all the required parameter in configfile\n")
			os.Exit(1)
		}

		fmt.Printf("Backend resource will be created in project %v\n", projectID)
		err := createInstance(projectID, zone, instanceName, machineType, sourceImage, networkName)
		if err != nil {
			fmt.Printf("Unable to create instance: %v\n", err)
			os.Exit(1)
		}
		err = createBucketClassLocation(projectID, bucketName)
		if err != nil {
			fmt.Printf("Unable to create bucket: %v\n", err)
			os.Exit(1)
		}
		err = createDataset(projectID, datasetID)
		if err != nil {
			fmt.Printf("Unable to create dataset: %v\n", err)
			os.Exit(1)
		}
		err = createTableExplicitSchema(projectID, datasetID)
		if err != nil {
			fmt.Printf("Unable to create table: %v\n", err)
			os.Exit(1)
		}

		viper.Set("bucketname", bucketName)
		err = viper.WriteConfig()
		if err != nil {
			fmt.Printf("Unable to save config: %v\n", err)
			os.Exit(1)
		}

		err = createFirewallRule(projectID, firewallRuleName, networkName)
		if err != nil {
			fmt.Printf("Unable to create firewall rule: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// createInstance sends an instance creation request to the Compute Engine API and waits for it to complete.
func createInstance(projectID, zone, instanceName, machineType, sourceImage, networkName string) error {
	// projectID := "your_project_id"
	// zone := "europe-central2-b"
	// instanceName := "your_instance_name"
	// machineType := "n1-standard-1"
	// sourceImage := "projects/debian-cloud/global/images/family/debian-10"
	// networkName := "global/networks/default"

	ctx := context.Background()
	projectsClient, err := resourcemanager.NewProjectsClient(ctx)
	if err != nil {
		return fmt.Errorf("NewProjectClient: %v", err)
	}
	defer projectsClient.Close()

	projectreq := &resourcemanagerpb.GetProjectRequest{
		// TODO: Fill request struct fields.
		// See https://pkg.go.dev/google.golang.org/genproto/googleapis/cloud/resourcemanager/v3#GetProjectRequest.
		Name: *proto.String("projects/" + projectID),
	}
	projectresp, err := projectsClient.GetProject(ctx, projectreq)
	if err != nil {
		// TODO: Handle error.
		return fmt.Errorf("unable to get projectnumber: %v", err)
	}
	projectNum := strings.Split(projectresp.Name, "/")[1]
	fmt.Printf("Project Number is %v\n", projectNum)

	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("NewInstancesRESTClient: %v", err)
	}
	defer instancesClient.Close()

	script := `#! /bin/bash
apt update
apt install python3-pip -y
curl -fsSL https://get.pulumi.com | sh
cp -R .pulumi/bin/* /usr/local/bin/
mkdir /home/pulumidemobackend
chmod +777 -R /home/pulumidemobackend
cat <<EOF > requirements.txt
fastapi==0.82.0
google-api-core==2.10.0
google-auth==2.11.0
google-cloud-bigquery==3.3.2
google-cloud-bigquery-storage==2.15.0
google-cloud-compute==1.5.2
google-cloud-core==2.3.2
google-crc32c==1.5.0
google-resumable-media==2.3.3
googleapis-common-protos==1.56.4
proto-plus==1.22.1
protobuf==4.21.5
pulumi==3.39.3
pulumi-gcp==6.36.0
pydantic==1.10.2
typing==3.7.4.3
typing-extensions==4.3.0
uvicorn==0.13.3
EOF
pip3 install -r requirements.txt
`

	req := &computepb.InsertInstanceRequest{
		Project: projectID,
		Zone:    zone,
		InstanceResource: &computepb.Instance{
			Name: proto.String(instanceName),
			Disks: []*computepb.AttachedDisk{
				{
					InitializeParams: &computepb.AttachedDiskInitializeParams{
						DiskSizeGb:  proto.Int64(10),
						SourceImage: proto.String(sourceImage),
					},
					AutoDelete: proto.Bool(true),
					Boot:       proto.Bool(true),
					Type:       proto.String(computepb.AttachedDisk_PERSISTENT.String()),
				},
			},
			MachineType: proto.String(fmt.Sprintf("zones/%s/machineTypes/%s", zone, machineType)),
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					Name: proto.String(networkName),
					AccessConfigs: []*computepb.AccessConfig{
						{
							NetworkTier: proto.String("PREMIUM"),
						},
					},
				},
			},
			Metadata: &computepb.Metadata{
				Items: []*computepb.Items{
					{
						Key:   proto.String("startup-script"),
						Value: proto.String(script),
					},
				},
			},
			ServiceAccounts: []*computepb.ServiceAccount{
				{
					Email: proto.String(projectNum + "-compute@developer.gserviceaccount.com"),
					Scopes: []string{
						"https://www.googleapis.com/auth/cloud-platform",
					},
				},
			},
		},
	}

	op, err := instancesClient.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to create instance: %v", err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("unable to wait for the operation: %v", err)
	}

	fmt.Println("Instance created")

	return nil
}

func createBucketClassLocation(projectID, bucketName string) error {
	// projectID := "my-project-id"
	// bucketName := "bucket-name"
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	storageClassAndLocation := &storage.BucketAttrs{
		StorageClass: "STANDARD",
		Location:     "US",
	}
	bucket := client.Bucket(bucketName)
	if err := bucket.Create(ctx, projectID, storageClassAndLocation); err != nil {
		return fmt.Errorf("Bucket(%q).Create: %v", bucketName, err)
	}
	fmt.Printf("Created bucket %v in %v with storage class %v\n", bucketName, storageClassAndLocation.Location, storageClassAndLocation.StorageClass)

	// Upload an object with storage.Writer.
	object := "app_file/"
	b := []byte("")
	buf := bytes.NewBuffer(b)
	wc := client.Bucket(bucketName).Object(object).NewWriter(ctx)
	wc.ChunkSize = 0 // note retries are not supported for chunk size 0.

	if _, err = io.Copy(wc, buf); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	// Data can continue to be added to the file until the writer is closed.
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	fmt.Printf("%v uploaded to %v.\n", object, bucketName)

	return nil
}

// createDataset demonstrates creation of a new dataset using an explicit destination location.
func createDataset(projectID, datasetID string) error {
	// projectID := "my-project-id"
	// datasetID := "mydataset"
	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	meta := &bigquery.DatasetMetadata{
		Location: "US", // See https://cloud.google.com/bigquery/docs/locations
	}
	if err := client.Dataset(datasetID).Create(ctx, meta); err != nil {
		return err
	}
	fmt.Printf("Created dataset %v in %v\n", datasetID, meta.Location)
	return nil
}

// createTableExplicitSchema demonstrates creating a new BigQuery table and specifying a schema.
func createTableExplicitSchema(projectID, datasetID string) error {
	// projectID := "my-project-id"
	// datasetID := "mydatasetid"
	// tableID := "mytableid"
	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	tableID := "job-status"

	sampleSchema := bigquery.Schema{
		{Name: "job_name", Type: bigquery.StringFieldType},
		{Name: "job_status", Type: bigquery.StringFieldType},
		{Name: "message", Type: bigquery.StringFieldType},
	}

	metaData := &bigquery.TableMetadata{
		Schema: sampleSchema,
		// ExpirationTime: time.Now().AddDate(1, 0, 0), // Table will be automatically deleted in 1 year.
	}
	tableRef := client.Dataset(datasetID).Table(tableID)
	if err := tableRef.Create(ctx, metaData); err != nil {
		return err
	}
	fmt.Printf("Created table %v under dataset %v\n", tableID, datasetID)

	tableID = "job-result"

	sampleSchema = bigquery.Schema{
		{Name: "job_name", Type: bigquery.StringFieldType},
		{Name: "hostname", Type: bigquery.StringFieldType},
		{Name: "os_version", Type: bigquery.StringFieldType},
		{Name: "status", Type: bigquery.StringFieldType},
		{Name: "result", Type: bigquery.StringFieldType},
		{Name: "message", Type: bigquery.StringFieldType},
	}

	metaData = &bigquery.TableMetadata{
		Schema: sampleSchema,
		// ExpirationTime: time.Now().AddDate(1, 0, 0), // Table will be automatically deleted in 1 year.
	}
	tableRef = client.Dataset(datasetID).Table(tableID)
	if err := tableRef.Create(ctx, metaData); err != nil {
		return err
	}
	fmt.Printf("Created table %v under dataset %v\n", tableID, datasetID)

	return nil
}

// createFirewallRule creates a firewall rule allowing for incoming HTTP:8080 access from the entire Internet.
func createFirewallRule(projectID, firewallRuleName, networkName string) error {
	// projectID := "your_project_id"
	// firewallRuleName := "europe-central2-b"
	// networkName := "global/networks/default"

	ctx := context.Background()
	firewallsClient, err := compute.NewFirewallsRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("NewInstancesRESTClient: %v", err)
	}
	defer firewallsClient.Close()

	firewallRule := &computepb.Firewall{
		Allowed: []*computepb.Allowed{
			{
				IPProtocol: proto.String("tcp"),
				Ports:      []string{"8080"},
			},
		},
		Direction:   proto.String(computepb.Firewall_INGRESS.String()),
		Name:        &firewallRuleName,
		TargetTags:  []string{},
		Network:     &networkName,
		Description: proto.String("Allowing TCP traffic on port 80 and 443 from Internet."),
	}

	// Note that the default value of priority for the firewall API is 1000.
	// If you check the value of `firewallRule.GetPriority()` at this point it
	// will be equal to 0, however it is not treated as "set" by the library and thus
	// the default will be applied to the new rule. If you want to create a rule that
	// has priority == 0, you need to explicitly set it so:

	// firewallRule.Priority = proto.Int32(0)

	req := &computepb.InsertFirewallRequest{
		Project:          projectID,
		FirewallResource: firewallRule,
	}

	op, err := firewallsClient.Insert(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to create firewall rule: %v", err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("unable to wait for the operation: %v", err)
	}

	fmt.Println("Firewall rule created")

	return nil
}

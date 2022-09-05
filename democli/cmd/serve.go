/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

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
		bucketName = viper.GetString("bucketname")
		datasetID = viper.GetString("datasetid")
		tableID = viper.GetString("tableid")

		if projectID == "" || zone == "" || instanceName == "" || bucketName == "" || datasetID == "" || tableID == "" {
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
		err = createTableExplicitSchema(projectID, datasetID, tableID)
		if err != nil {
			fmt.Printf("Unable to create table: %v\n", err)
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
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("NewInstancesRESTClient: %v", err)
	}
	defer instancesClient.Close()

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
func createTableExplicitSchema(projectID, datasetID, tableID string) error {
	// projectID := "my-project-id"
	// datasetID := "mydatasetid"
	// tableID := "mytableid"
	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	sampleSchema := bigquery.Schema{
		{Name: "full_name", Type: bigquery.StringFieldType},
		{Name: "age", Type: bigquery.IntegerFieldType},
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
	return nil
}

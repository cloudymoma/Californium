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
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean all the backend resource",
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
		// tableID = viper.GetString("tableid")

		if projectID == "" || zone == "" || instanceName == "" || bucketName == "" || datasetID == "" {
			fmt.Printf("Please give all the required parameter in configfile\n")
			os.Exit(1)
		}

		fmt.Printf("Backend resource will be deleted in project %v\n", projectID)
		err := deleteInstance(projectID, zone, instanceName)
		if err != nil {
			fmt.Printf("Unable to delete instance: %v\n", err)
			os.Exit(1)
		}
		err = deleteBucket(bucketName)
		if err != nil {
			fmt.Printf("Unable to delete bucket: %v\n", err)
			os.Exit(1)
		}
		err = deleteDataset(projectID, datasetID)
		if err != nil {
			fmt.Printf("Unable to delete dataset: %v\n", err)
			os.Exit(1)
		}
		err = deleteFirewallRule(projectID, firewallRuleName)
		if err != nil {
			fmt.Printf("Unable to delete firewall rule: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cleanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cleanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// deleteInstance sends a delete request to the Compute Engine API and waits for it to complete.
func deleteInstance(projectID, zone, instanceName string) error {
	// projectID := "your_project_id"
	// zone := "europe-central2-b"
	// instanceName := "your_instance_name"
	ctx := context.Background()
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("NewInstancesRESTClient: %v", err)
	}
	defer instancesClient.Close()

	req := &computepb.DeleteInstanceRequest{
		Project:  projectID,
		Zone:     zone,
		Instance: instanceName,
	}

	op, err := instancesClient.Delete(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to delete instance: %v", err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("unable to wait for the operation: %v", err)
	}

	fmt.Printf("Instance deleted\n")

	return nil
}

// deleteBucket deletes the bucket.
func deleteBucket(bucketName string) error {
	// bucketName := "bucket-name"
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	bucket := client.Bucket(bucketName)
	if err := bucket.Delete(ctx); err != nil {
		return fmt.Errorf("Bucket(%q).Delete: %v", bucketName, err)
	}
	fmt.Printf("Bucket %v deleted\n", bucketName)
	return nil
}

// deleteDataset demonstrates the deletion of an empty dataset.
func deleteDataset(projectID, datasetID string) error {
	// projectID := "my-project-id"
	// datasetID := "mydataset"
	ctx := context.Background()

	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	// To recursively delete a dataset and contents, use DeleteWithContents.
	if err := client.Dataset(datasetID).DeleteWithContents(ctx); err != nil {
		return fmt.Errorf("delete: %v", err)
	}
	fmt.Printf("Dataset %v deleted\n", datasetID)
	return nil
}

// deleteFirewallRule deletes a firewall rule from the project.
func deleteFirewallRule(projectID, firewallRuleName string) error {
	// projectID := "your_project_id"
	// firewallRuleName := "europe-central2-b"

	ctx := context.Background()
	firewallsClient, err := compute.NewFirewallsRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("NewInstancesRESTClient: %v", err)
	}
	defer firewallsClient.Close()

	req := &computepb.DeleteFirewallRequest{
		Project:  projectID,
		Firewall: firewallRuleName,
	}

	op, err := firewallsClient.Delete(ctx, req)
	if err != nil {
		return fmt.Errorf("unable to delete firewall rule: %v", err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("unable to wait for the operation: %v", err)
	}

	fmt.Println("Firewall rule deleted")

	return nil
}

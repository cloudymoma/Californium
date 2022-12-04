/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var jobFile string

var (
	projectID    string
	zone         string
	instanceName string
	machineType  string
	sourceImage  string
	networkName  string
	bucketName   string
	datasetID    string
	serverIP     string
	serverPort   string
	// tableID      string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "democli",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.democli.yaml)")
	cobra.OnInitialize(initConfig)
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.cobra.json)")
	rootCmd.PersistentFlags().StringVarP(&projectID, "projectid", "p", "", "config gcp projectID")
	rootCmd.PersistentFlags().StringVarP(&zone, "zone", "z", "", "config gce zone (available zone list https://cloud.google.com/compute/docs/regions-zones?hl=zh-cn#available)")
	rootCmd.PersistentFlags().StringVarP(&instanceName, "instancename", "i", "", "config gce instance name")
	rootCmd.PersistentFlags().StringVarP(&machineType, "machinetype", "m", "n1-standard-1", "config gce instance machine type (default is n1-standard-1, available machine type list https://cloud.google.com/compute/docs/general-purpose-machines)")
	rootCmd.PersistentFlags().StringVarP(&sourceImage, "sourceimage", "s", "projects/debian-cloud/global/images/family/debian-11", "config gce instance image (default is projects/debian-cloud/global/images/family/debian-11)")
	rootCmd.PersistentFlags().StringVarP(&networkName, "networkname", "n", "global/networks/default", "config gce network (default is global/networks/default)")
	rootCmd.PersistentFlags().StringVarP(&bucketName, "bucketName", "b", "gametest", "config gcs backetname")
	rootCmd.PersistentFlags().StringVarP(&datasetID, "datasetid", "d", "gametest", "config bigquery dataset name")
	rootCmd.PersistentFlags().StringVarP(&serverIP, "serverip", "a", "127.0.0.1", "config backend server ip address")
	rootCmd.PersistentFlags().StringVarP(&serverPort, "serverport", "l", "8080", "config backend server listen port")
	rootCmd.PersistentFlags().StringVarP(&jobFile, "jobconfig", "j", "job.json", "config job config file path")
	// rootCmd.PersistentFlags().StringVarP(&tableID, "tableid", "t", "", "config bigquery table name")
	viper.BindPFlag("projectid", rootCmd.PersistentFlags().Lookup("projectid"))
	viper.BindPFlag("zone", rootCmd.PersistentFlags().Lookup("zone"))
	viper.BindPFlag("instanceName", rootCmd.PersistentFlags().Lookup("instanceName"))
	viper.BindPFlag("machinetype", rootCmd.PersistentFlags().Lookup("machinetype"))
	viper.BindPFlag("sourceimage", rootCmd.PersistentFlags().Lookup("sourceimage"))
	viper.BindPFlag("networkname", rootCmd.PersistentFlags().Lookup("networkname"))
	viper.BindPFlag("bucketName", rootCmd.PersistentFlags().Lookup("bucketName"))
	viper.BindPFlag("datasetid", rootCmd.PersistentFlags().Lookup("datasetid"))
	// viper.BindPFlag("tableid", rootCmd.PersistentFlags().Lookup("tableid"))
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("json")
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Please setup the configFile first")
		os.Exit(1)
	} else {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

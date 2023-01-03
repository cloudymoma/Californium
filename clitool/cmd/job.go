/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
)

type osList struct {
	Win7  string `json:"win7,omitempty"`
	Win8  string `json:"win8,omitempty"`
	Win10 string `json:"win10,omitempty"`
	Win11 string `json:"win11,omitempty"`
}

type jobConfig struct {
	Job_name      string `json:"job_name"`
	Project_id    string `json:"project_id"`
	Region        string `json:"region"`
	Zone          string `json:"zone"`
	Instance_type string `json:"instance_type"`
	Vpc_network   string `json:"vpc_network"`
	Gcp_bucket    string `json:"gcp_bucket"`
	Os_list       osList `json:"os_list"`
	Cmd           string `json:"cmd"`
	App_name      string `json:"app_name"`
	App_file_name string `json:"app_file_name"`
}

// jobCmd represents the job command
var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("job called")
	// },
}

func init() {
	rootCmd.AddCommand(jobCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// jobCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// jobCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

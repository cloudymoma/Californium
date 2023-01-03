/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		projectID = viper.GetString("projectid")
		zone = viper.GetString("zone")
		machineType = viper.GetString("machinetype")
		bucketName = viper.GetString("bucketname")
		networkName = viper.GetString("networkname")

		// Open our jsonFile
		jsonFile, err := os.Open(jobFile)
		// if we os.Open returns an error then handle it
		if err != nil {
			fmt.Printf("Unable to open job config file: %v\n", err)
			os.Exit(1)
		}

		defer jsonFile.Close()

		// read our opened jsonFile as a byte array.
		byteValue, _ := ioutil.ReadAll(jsonFile)

		jobconfig := jobConfig{
			Job_name:      "",
			Instance_type: "",
			Vpc_network:   "",
			Project_id:    "",
			Gcp_bucket:    "",
			Region:        "",
			Zone:          "",
			Os_list:       osList{Win7: "", Win8: "", Win10: "", Win11: ""},
			Cmd:           "",
			App_name:      "",
			App_file_name: "",
		}
		// we unmarshal our byteArray which contains our
		// jsonFile's content into 'users' which we defined above
		json.Unmarshal(byteValue, &jobconfig)

		if jobconfig.Job_name == "" {
			fmt.Println("Please provide job name in jobFile!")
			os.Exit(1)
		}

		if jobconfig.Instance_type == "" {
			jobconfig.Instance_type = machineType
		}

		if jobconfig.Vpc_network == "" {
			jobconfig.Vpc_network = networkName
		}

		if jobconfig.Project_id == "" {
			jobconfig.Project_id = projectID
		}

		if jobconfig.Gcp_bucket == "" {
			jobconfig.Gcp_bucket = bucketName
		}

		if jobconfig.Zone == "" {
			jobconfig.Zone = zone
		}

		if jobconfig.Region == "" {
			jobconfig.Region = strings.Split(jobconfig.Zone, "-")[0] + "-" + strings.Split(jobconfig.Zone, "-")[1]
		}

		err = deleteJob(jobconfig, serverIP, serverPort)

		if err != nil {
			fmt.Printf("Delete Job failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	jobCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func deleteJob(jobconfig jobConfig, server string, port string) error {
	url := "http://" + server + ":" + port + "/jobs/"

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(jobconfig)
	if err != nil {
		return fmt.Errorf("job config load fail: %v", err)
	}

	req, _ := http.NewRequest("DELETE", url, &buf)

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return fmt.Errorf("delete job failed: %v", err)
	}

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return fmt.Errorf("delete job failed: %v", err)
	}

	fmt.Printf("RESPONSE:\n%s", string(respDump))

	return nil
}

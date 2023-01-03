/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/spf13/cobra"
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
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

		err = queryJob(jobconfig, serverIP, serverPort)

		if err != nil {
			fmt.Printf("Query Job failed: %v\n", err)
			os.Exit(1)
		}

	},
}

func init() {
	jobCmd.AddCommand(queryCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// queryCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// queryCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func queryJob(jobconfig jobConfig, server string, port string) error {
	url := "http://" + server + ":" + port + "/jobs/" + jobconfig.Job_name

	req, _ := http.NewRequest("GET", url, nil)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return fmt.Errorf("query job failed: %v", err)
	}

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return fmt.Errorf("query job failed: %v", err)
	}

	fmt.Printf("RESPONSE:\n%s", string(respDump))

	return nil
}

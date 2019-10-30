/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	serverUrl string
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		loginForm := make(map[string]interface{})
		email := promptui.Prompt{
			Label: "Email",
		}
		emailResult, _ := email.Run()
		loginForm["email"] = emailResult
		password := promptui.Prompt{
			Label: "Password",
			Mask:  '*',
		}
		passwordResult, _ := password.Run()
		loginForm["password"] = passwordResult
		httpClient := &http.Client{}
		jsonData, _ := json.Marshal(loginForm)
		response, _ := httpClient.Post(serverUrl+"/v1/login", "application/json", bytes.NewReader(jsonData))
		fmt.Printf("response = %v", response)
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	loginCmd.Flags().StringVarP(&serverUrl, "target", "t", "http://localhost:3000", "used for setting the api target")
}

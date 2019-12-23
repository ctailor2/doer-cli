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
	"net/http"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"bufio"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var signupCmd = &cobra.Command{
	Use:   "signup",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		resourcesResponse := getBaseResources(Link{Href: viper.GetString("server-url") + "/v1/"})
		scanner := bufio.NewScanner(os.Stdin)
		signup(scanner, resourcesResponse.Links["signup"].Href)
	},
}

func signup(scanner *bufio.Scanner, url string) {
	form := make(map[string]interface{})
	fmt.Print("Email: ")
	scanner.Scan()
	emailResult := scanner.Text()
	form["email"] = emailResult
	fmt.Print("Password: ")
	scanner.Scan()
	passwordResult := scanner.Text()
	fmt.Print("Password Confirmation: ")
	scanner.Scan()
	passwordConfirmationResult := scanner.Text()
	if passwordResult != passwordConfirmationResult {
		fmt.Println("Password confirmation and password do not match.")
		return
	}
	form["password"] = passwordResult
	httpClient := &http.Client{}
	jsonData, _ := json.Marshal(form)
	response, _ := httpClient.Post(url, "application/json", bytes.NewReader(jsonData))
	var SessionResponse SessionResponse
	jsonParseErr := json.NewDecoder(response.Body).Decode(&SessionResponse)
	if jsonParseErr != nil {
		fmt.Println(jsonParseErr)
	}
	viper.Set("session-token", SessionResponse.Session.Token)
	viper.Set("root-href", SessionResponse.Links["root"].Href)
	err := viper.WriteConfig()
	if err != nil {
		fmt.Println(err)
	}
}

func init() {
	rootCmd.AddCommand(signupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

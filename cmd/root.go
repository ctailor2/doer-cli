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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	serverUrl string
)

type ResourcesResponse struct {
	Links map[string]Link `json:"_links"`
}

type Link struct {
	Href string `json:"href"`
}

type SessionResponse struct {
	Session Session         `json:"session"`
	Links   map[string]Link `json:"_links"`
}

type Session struct {
	Token string `json:"token"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "doer-cli",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		resourcesResponse := getBaseOrRootResourcesResponse()
		scanner := bufio.NewScanner(os.Stdin)
		action := chooseNextAction(resourcesResponse, scanner)
		switch action {
		case "login":
			login(scanner, resourcesResponse.Links[action].Href)
		case "signup":
			signup(scanner, resourcesResponse.Links[action].Href)
		default:
			fmt.Println("Chosen selection has not yet been implemented")
		}
	},
}

func chooseNextAction(resourcesResponse ResourcesResponse, scanner *bufio.Scanner) string {
	resourceOptions := make([]string, 0, len(resourcesResponse.Links))
	for k := range resourcesResponse.Links {
		if k != "self" {
			resourceOptions = append(resourceOptions, k)
		}
	}
	sort.Strings(resourceOptions)
	fmt.Printf("Choose action %v: ", resourceOptions)
	scanner.Scan()
	return scanner.Text()
}

func getBaseOrRootResourcesResponse() ResourcesResponse {
	client := &http.Client{}
	var response *http.Response
	if viper.IsSet("session-token") {
		href, _ := url.Parse(viper.GetString("root-href"))
		req, _ := http.NewRequest("GET", href.String(), nil)
		req.Header.Add("Session-Token", viper.GetString("session-token"))
		response, _ = client.Do(req)
	} else {
		response, _ = client.Get(serverUrl + "/v1/")
	}
	var resourcesResponse ResourcesResponse
	jsonParseErr := json.NewDecoder(response.Body).Decode(&resourcesResponse)
	if jsonParseErr != nil {
		fmt.Println(jsonParseErr)
	}
	return resourcesResponse
}

func login(scanner *bufio.Scanner, url string) {
	form := make(map[string]interface{})
	fmt.Print("Email: ")
	scanner.Scan()
	emailResult := scanner.Text()
	form["email"] = emailResult
	fmt.Print("Password: ")
	scanner.Scan()
	passwordResult := scanner.Text()
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.doer-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringVarP(&serverUrl, "api", "a", "http://localhost:8080", "used for setting the api target")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".doer-cli" (without extension).
		viper.AddConfigPath(home)
		configName := ".doer-cli"
		viper.SetConfigName(configName)
		cfgFile = home + "/" + configName + ".yml"
	}

	_, err := os.Stat(cfgFile)
	if os.IsNotExist(err) {
		var file, err = os.Create(cfgFile)
		if err != nil {
			fmt.Println("Error creating config file: ", cfgFile)
			fmt.Println(err)
		}
		defer file.Close()
	}
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	viper.Set("server-url", serverUrl)
	err = viper.WriteConfig()
	if err != nil {
		fmt.Println(err)
	}
}

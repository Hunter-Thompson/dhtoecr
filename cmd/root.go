/*
Copyright © 2022 Hunter Thompson aatman@auroville.org.in
*/
package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	configFile   string
	awsRegion    string
	awsAccountID string
	client       *ecr.Client
	dryRun       bool
)

type config struct {
	RegistryMap map[string][]string `yaml:"registryMap" json:"registryMap"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dh2ecr",
	Short: "Copies over images from dockerhub to ecr",

	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		checkVars()
		if !dryRun {
			getEcrClient()
			loginToEcr()
		}
		cfg := setConfig()

		for registryName, imageTags := range cfg.RegistryMap {
			pullAndPush(registryName, imageTags, dryRun)
		}
	},
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

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dh2ecr.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	rootCmd.Flags().StringVarP(&awsRegion, "aws-region", "r", "", "aws region")
	rootCmd.Flags().StringVarP(&awsAccountID, "aws-account-id", "a", "", "aws account id")
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "dry run")
}

// func to login to ecr
func getEcrClient() {
	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(awsRegion))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	client = ecr.NewFromConfig(cfg)
}

func loginToEcr() {
	// get auth token
	authToken, err := client.GetAuthorizationToken(context.TODO(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		log.Fatalf("unable to get auth token, %v", err)
	}

	// login to ecr registry using docker command
	token, err := base64.StdEncoding.DecodeString(*authToken.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		log.Fatalf("unable to decode auth token, %v", err)
	}
	endpoint := authToken.AuthorizationData[0].ProxyEndpoint

	if len(token) == 0 {
		log.Fatal("empty auth token")
	}

	tokenSlice := strings.Split(string(token), ":")
	if len(tokenSlice) != 2 {
		log.Fatal("invalid auth token")
	}

	// login to ecr
	cmd := exec.Command("docker", "login", "-u", tokenSlice[0], "-p", tokenSlice[1], *endpoint)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("unable to login to ecr, %v", err)
	}
}

func setConfig() *config {
	log.Println("reading config", configFile)
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("unable to read config file, %v", err)
	}

	var cfg config
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		log.Fatalf("unable to unmarshal config file, %v", err)
	}

	log.Println("config:", cfg)

	return &cfg
}

func pullAndPush(registryName string, imageTags []string, dry bool) {
	if !dry {
		_, err := client.CreateRepository(context.TODO(), &ecr.CreateRepositoryInput{
			RepositoryName: &registryName,
		})
		if err != nil {
			if err.Error() != "RepositoryAlreadyExistsException" {
				log.Fatalf("unable to create repository, %v", err)
			}
			log.Printf("repository %s already exists", registryName)
		}
	}

	for _, image := range imageTags {
		imageSlice := strings.Split(image, ":")
		if len(imageSlice) != 2 {
			log.Fatalf("invalid image tag %s", image)
		}

		tag := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", awsAccountID, awsRegion, image)
		log.Printf("pulling image %s from docker hub", image)
		if !dry {
			cmd := exec.Command("docker", "pull", image)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				log.Fatalf("unable to pull image %s, %v", image, err)
			}
		}

		log.Printf("tagging image %s with tag %s", image, tag)
		if !dry {
			cmd := exec.Command("docker", "tag", image, tag)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				log.Fatalf("unable to tag image %s, %v", image, err)
			}
		}

		log.Printf("pushing image %s to ecr", tag)
		if !dry {
			cmd := exec.Command("docker", "push", tag)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				log.Fatalf("unable to push image %s, %v", tag, err)
			}
		}
	}
}

func checkVars() {
	if awsRegion == "" {
		log.Fatal("aws region is not set")
	}

	if awsAccountID == "" {
		log.Fatal("aws account id is not set")
	}

	if configFile == "" {
		log.Fatal("config file is not set")
	}
}

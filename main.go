package main

import (
	"context"
	"encoding/base64"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v56/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func main() {
	log.Println("Starting Terraform Cloud/Enterprise GitHub PR comments Run Tasks...")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on HTTP port: %s", port)

	token := os.Getenv("GITHUB_OAUTH_TOKEN")
	ghAppID := os.Getenv("GITHUB_APP_ID")
	ghAppKey := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	ghAppInstallationID := os.Getenv("GITHUB_APP_INSTALLATION_ID")
	if token == "" && (ghAppID == "" || ghAppKey == "" || ghAppInstallationID == "") {
		log.Fatal("Missing an authentication config for GitHub")
	}

	var (
		ghClient        *github.Client
		ghGraphQLClient *githubv4.Client
	)
	if token != "" {
		sts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		c := oauth2.NewClient(context.Background(), sts)
		ghClient = github.NewClient(c)
		ghGraphQLClient = githubv4.NewClient(c)
	} else {
		appID, err := strconv.ParseInt(ghAppID, 10, 64)
		if err != nil {
			log.Fatalf("Invalid GitHub App id: %v", err)
		}
		installationID, err := strconv.ParseInt(ghAppInstallationID, 10, 64)
		if err != nil {
			log.Fatalf("Invalid GitHub App installation id: %v", err)
		}
		key, err := base64.StdEncoding.DecodeString(ghAppKey)
		if err != nil {
			log.Fatalf("Failed to decode GitHub App private key: %v", err)
		}
		itr, err := ghinstallation.New(http.DefaultTransport, appID, installationID, key)
		if err != nil {
			log.Fatalf("Failed to create GitHub client authenticated by GitHub App: %v", err)
		}
		c := &http.Client{Transport: itr}
		ghClient = github.NewClient(c)
		ghGraphQLClient = githubv4.NewClient(c)
	}

	handler := newHandler(ghClient, ghGraphQLClient)
	http.HandleFunc("/", handler.handleRunTask)

	log.Fatal(http.ListenAndServe(net.JoinHostPort("", port), nil))
}

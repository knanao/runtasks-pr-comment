package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v56/github"
	"github.com/shurcooL/githubv4"
)

type handler struct {
	ghClient        *github.Client
	ghGraphQLClient *githubv4.Client
	httpClient      *http.Client
}

func newHandler(ghClient *github.Client, ghGraphQLClient *githubv4.Client) *handler {
	return &handler{
		ghClient:        ghClient,
		ghGraphQLClient: ghGraphQLClient,
		httpClient:      &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *handler) handleRunTask(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to load request: %v", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	key := os.Getenv("TFC_RUN_TASK_HMAC_KEY")
	mac := hmac.New(sha512.New, []byte(key))
	mac.Write(body)
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	requestedSig := r.Header.Get("X-TFC-Task-Signature")
	if expectedSig != requestedSig {
		log.Printf("Invalid x-tfc-task-signature value: %s. Please check your HMAC Key", requestedSig)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		log.Printf("This method is not alloed: %s. Expected: %s.", r.Method, http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req := &TFERunTasksRequest{}
	if err := json.Unmarshal(body, req); err != nil {
		log.Printf("Failed to unmarshal request: %v", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.AccessToken == "test-token" {
		log.Printf("Succeeded initializing run tasks")
		return
	}

	ctx := context.Background()
	if req.VCSPullRequestURL == "" {
		log.Printf("Skip this run because this might not be the event based on PR: %s", req.RunID)

		msg := "Skipped pushing the plan result to VCS"
		if err := h.sendCallback(ctx, req.TaskResultCallbackURL, req.AccessToken, msg); err != nil {
			log.Printf("Failed to send callback to TFC: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	plan, err := parsePlan(ctx, h.httpClient, req.PlanJSONAPIURL, req.AccessToken)
	if err != nil {
		log.Printf("Failed to get the plan: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	url, err := newGitURL(req.VCSPullRequestURL)
	if err != nil {
		log.Printf("Unable to parse VCS pull request URL: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var (
		owner    = url.Owner()
		repo     = url.Repository()
		prNumber = url.PullRequest()
	)

	latestComment, err := findLatestComment(ctx, h.ghGraphQLClient, owner, repo, prNumber)
	if err != nil && !errors.Is(err, errNotFound) {
		log.Printf("Unable to query the previous comment to minimize: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	comment, err := makeIssueComment(plan, req.RunAppURL, req.VCSCommitURL)
	if err != nil {
		log.Printf("Failed to get the plan: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := createIssueComment(ctx, h.ghClient, owner, repo, prNumber, comment); err != nil {
		log.Printf("Failed to create an issue comment: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if latestComment != nil && bool(!latestComment.IsMinimized) {
		if err := minimizeComment(ctx, h.ghGraphQLClient, latestComment.ID, "OUTDATED"); err != nil {
			log.Printf("Failed to minimize comment: %v", err)
			return
		}
	}

	msg := "Succeeded pushing the plan result to VCS"
	if err := h.sendCallback(ctx, req.TaskResultCallbackURL, req.AccessToken, msg); err != nil {
		log.Printf("Failed to send callback to TFC: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *handler) sendCallback(ctx context.Context, url, token, message string) error {
	data := &TFERunTasksResponse{
		Data: &TFERunTasksResponseData{
			Type: "task-results",
			Attributes: &TFERunTasksResponseAttributes{
				Status:  "passed",
				Message: message,
			},
		},
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		log.Printf("Failed to marshal a response data")
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		log.Printf("Unable to send data to webhook url: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Unexpected status was returned: %d", resp.StatusCode)
		return fmt.Errorf("Unexpected status was returned: %d", resp.StatusCode)
	}

	_, err = io.ReadAll(resp.Body)
	return err
}

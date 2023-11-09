package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v56/github"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shurcooL/githubv4"
)

func makeIssueComment(plan *tfjson.Plan, runURL, commitURL string) (string, error) {
	const (
		tasksBadgeURL = `<!-- runtasks-pr-comment -->
[![RUN_TASKS](https://img.shields.io/static/v1?label=TFE&message=Run_Tasks&color=success&style=flat)](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/settings/run-tasks)`
		runBadgeURL   = `[![RUNS](https://img.shields.io/static/v1?label=TFE&message=Run&style=flat)](%s)`
		description   = `This run task was triggered by %s.`
		title         = `### Terraform Cloud/Enterprise Plan Output`
		noChanges     = "```\nNo changes. Your infrastructure matches the configuration.\n```"
		changeDetails = "<details>\n<summary>%s</summary>\n\n```go\n%s\n```\n</details>"
	)

	var b strings.Builder
	b.WriteString(tasksBadgeURL)

	fmt.Fprintf(&b, " ")
	fmt.Fprintf(&b, runBadgeURL, runURL)
	b.WriteString("\n\n")

	fmt.Fprintf(&b, description, commitURL)
	b.WriteString("\n\n")

	b.WriteString(title)
	b.WriteString("\n")

	changes := plan.ResourceChanges
	if len(changes) == 0 {
		b.WriteString(noChanges)
		return b.String(), nil
	}

	cs := &ChangeSummary{}
	var diff strings.Builder
	for _, c := range changes {
		if c.Change == nil {
			b.WriteString(noChanges)
			return b.String(), nil
		}

		if c.Change.Importing != nil {
			cs.Import++
			continue
		}

		for _, a := range c.Change.Actions {
			switch a {
			case tfjson.ActionCreate:
				cs.Add++
			case tfjson.ActionUpdate:
				cs.Change++
			case tfjson.ActionDelete:
				cs.Remove++
			}
		}

		action := UnmarshalActions(c.Change.Actions)
		if action == NoOp {
			continue
		}

		summary := fmt.Sprintf("%s %s", action.Symbol(), c.Address)
		detail := fmt.Sprintf(
			"%s %s %s",
			action.Symbol(),
			fmt.Sprintf("%s \"%s\" \"%s\"", "resource", c.Type, c.Name),
			cmp.Diff(c.Change.Before, c.Change.After),
		)
		diff.WriteString(fmt.Sprintf(changeDetails, summary, detail))
		diff.WriteString("\n\n")
	}

	b.WriteString(fmt.Sprintf("```\n%s\n```", cs.String()))
	b.WriteString("\n\n")

	b.WriteString(diff.String())

	outputs := plan.OutputChanges
	if len(outputs) > 0 {
		summary := fmt.Sprintf("Outputs %d planned to change", len(outputs))
		keys := make([]string, 0, len(plan.OutputChanges))
		for k := range plan.OutputChanges {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var ob strings.Builder
		for _, key := range keys {
			output := outputs[key]
			action := UnmarshalActions(output.Actions)
			if action != NoOp {
				detail := fmt.Sprintf(
					"%s %s: %s\n",
					action.Symbol(),
					key,
					cmp.Diff(output.Before, output.After),
				)
				ob.WriteString(detail)
			}
		}
		b.WriteString(fmt.Sprintf(changeDetails, summary, ob.String()))
	}

	return b.String(), nil
}

func createIssueComment(ctx context.Context, client *github.Client, owner, repo string, prNumber int, body string) error {
	_, _, err := client.Issues.CreateComment(
		ctx,
		owner,
		repo,
		prNumber,
		&github.IssueComment{
			Body: &body,
		},
	)
	return err
}

type issueCommentQuery struct {
	ID     githubv4.ID
	Author struct {
		Login githubv4.String
	}
	Body        githubv4.String
	IsMinimized githubv4.Boolean
}

type issueCommentsQuery struct {
	Nodes []issueCommentQuery
}

type pullRequestCommentQuery struct {
	Repository struct {
		PullRequest struct {
			Comments issueCommentsQuery `graphql:"comments(last: 100)"`
		} `graphql:"pullRequest(number: $prNumber)"`
	} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
}

var errNotFound = errors.New("not found")

func findLatestComment(ctx context.Context, client *githubv4.Client, owner, repo string, prNumber int) (*issueCommentQuery, error) {
	variables := map[string]interface{}{
		"repositoryOwner": githubv4.String(owner),
		"repositoryName":  githubv4.String(repo),
		"prNumber":        githubv4.Int(prNumber),
	}

	var q pullRequestCommentQuery
	if err := client.Query(ctx, &q, variables); err != nil {
		return nil, err
	}

	comment := filterLatestComment(q.Repository.PullRequest.Comments.Nodes)
	if comment == nil {
		return nil, errNotFound
	}
	return comment, nil
}

func filterLatestComment(comments []issueCommentQuery) *issueCommentQuery {
	const tag = "<!-- runtasks-pr-comment -->"

	for i := range comments {
		comment := comments[len(comments)-i-1]
		if strings.HasPrefix(string(comment.Body), tag) {
			return &comment
		}
	}
	return nil
}

type minimizeCommentMutation struct {
	MinimizeComment struct {
		MinimizedComment struct {
			IsMinimized bool
		}
	} `graphql:"minimizeComment(input: $input)"`
}

func minimizeComment(ctx context.Context, client *githubv4.Client, id githubv4.ID, classifier string) error {
	var m minimizeCommentMutation
	input := githubv4.MinimizeCommentInput{
		SubjectID:        id,
		Classifier:       githubv4.ReportedContentClassifiers(classifier),
		ClientMutationID: nil,
	}
	if err := client.Mutate(ctx, &m, input, nil); err != nil {
		return err
	}

	if !m.MinimizeComment.MinimizedComment.IsMinimized {
		return fmt.Errorf("cannot minimize comment. id: %s, classifier: %s", id, classifier)
	}
	return nil
}

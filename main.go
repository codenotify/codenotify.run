// Copyright 2022 Unknwon. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/flamego/flamego"
	"github.com/google/go-github/v45/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	log "unknwon.dev/clog/v2"

	"github.com/codenotify/codenotify.run/internal/conf"
)

func main() {
	if err := log.NewConsole(); err != nil {
		panic(err)
	}

	log.Info("Codenotify as a Service!")
	if conf.BuildTime != "" {
		log.Info("Build time: %s", conf.BuildTime)
		log.Info("Build commit: %s", conf.BuildCommit)
	}

	config, err := conf.Load()
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}

	f := flamego.Classic()
	f.Post("/-/webhook", func(r *http.Request) (int, string) {
		event := r.Header.Get("X-GitHub-Event")
		log.Trace("Received event: %s", event)

		if event != "pull_request" {
			return http.StatusOK, fmt.Sprintf("Event %q has been received but nothing to do", event)
		}

		var payload github.PullRequestEvent
		err = json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return http.StatusBadRequest, fmt.Sprintf("Failed to decode payload: %v", err)
		}
		if payload.Installation == nil || payload.Installation.ID == nil {
			return http.StatusBadRequest, "No installation or installation ID"
		} else if payload.Action == nil {
			return http.StatusBadRequest, "No action"
		}

		switch *payload.Action {
		case "opened":
			go handlePullRequestOpen(context.Background(), config.GitHubApp, &payload)
		case "synchronize":
			go handlePullRequestSynchronize(context.Background(), config.GitHubApp, &payload)
		default:
			return http.StatusOK, fmt.Sprintf("Event %q with action %q has been received but nothing to do", event, payload.Action)
		}
		return http.StatusAccepted, http.StatusText(http.StatusAccepted)
	})
	f.Run()
}

func newGitHubClient(ctx context.Context, appID, installationID int64, privateKey string) (*github.Client, error) {
	tr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, []byte(privateKey))
	if err != nil {
		return nil, errors.Wrap(err, "new transport")
	}

	client := github.NewClient(
		&http.Client{
			Transport: tr,
		},
	)

	token, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return nil, errors.Wrap(err, "create installation access token")
	}
	if token.Token == nil || *token.Token == "" {
		return nil, errors.New("empty token returned")
	}

	client = github.NewClient(
		oauth2.NewClient(
			ctx,
			oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: *token.Token,
				},
			),
		),
	)
	return client, nil
}

const commentMarker = `<!-- f05a7112-ce8b-4aaf-a203-f850b869431f -->`

func handlePullRequestOpen(ctx context.Context, githubApp conf.GitHubApp, payload *github.PullRequestEvent) {
	client, err := newGitHubClient(ctx, githubApp.AppID, *payload.Installation.ID, githubApp.PrivateKey)
	if err != nil {
		log.Error("Failed to create GitHub client: %v", err)
		return
	}

	comment, _, err := client.Issues.CreateComment(
		ctx,
		*payload.Repo.Owner.Login,
		*payload.Repo.Name,
		*payload.PullRequest.Number,
		&github.IssueComment{
			Body: github.String(commentMarker + "\n\n" + time.Now().Format(time.RFC3339)),
		},
	)
	if err != nil {
		log.Error("Failed to create comment on pull request %s: %v", payload.PullRequest.HTMLURL, err)
		return
	}
	log.Info("Created comment %s", *comment.HTMLURL)
}

func handlePullRequestSynchronize(ctx context.Context, githubApp conf.GitHubApp, payload *github.PullRequestEvent) {
	client, err := newGitHubClient(ctx, githubApp.AppID, *payload.Installation.ID, githubApp.PrivateKey)
	if err != nil {
		log.Error("Failed to create GitHub client: %v", err)
		return
	}

	// Iterate over first 100 comments on the pull request and update the previous
	// one. We don't look beyond 100 comments because it is very unlikely that the
	// previous comment is not within the first 100 comments.
	comments, _, err := client.Issues.ListComments(
		ctx,
		*payload.Repo.Owner.Login,
		*payload.Repo.Name,
		*payload.PullRequest.Number,
		&github.IssueListCommentsOptions{
			ListOptions: github.ListOptions{
				Page:    1,
				PerPage: 100,
			},
		},
	)
	if err != nil {
		log.Error("Failed to list comments on pull request %s: %v", payload.PullRequest.HTMLURL, err)
		return
	}

	commentBody := commentMarker + "\n\n" + time.Now().Format(time.RFC3339)
	for _, comment := range comments {
		if comment.Body == nil || !strings.Contains(*comment.Body, commentMarker) {
			continue
		}

		_, _, err = client.Issues.EditComment(
			ctx,
			*payload.Repo.Owner.Login,
			*payload.Repo.Name,
			*comment.ID,
			&github.IssueComment{
				Body: github.String(commentBody),
			},
		)
		if err != nil {
			log.Error("Failed to edit comment %s: %v", *comment.HTMLURL, err)
		}
		log.Info("Edited comment %s", *comment.HTMLURL)
		return
	}

	comment, _, err := client.Issues.CreateComment(
		ctx,
		*payload.Repo.Owner.Login,
		*payload.Repo.Name,
		*payload.PullRequest.Number,
		&github.IssueComment{
			Body: github.String(commentBody),
		},
	)
	if err != nil {
		log.Error("Failed to create comment on pull request %s: %v", payload.PullRequest.HTMLURL, err)
		return
	}
	log.Info("Created comment %s", *comment.HTMLURL)
}

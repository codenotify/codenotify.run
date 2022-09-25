// Copyright 2022 Unknwon. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v45/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	log "unknwon.dev/clog/v2"

	"github.com/codenotify/codenotify.run/internal/conf"
)

func newGitHubClient(ctx context.Context, appID, installationID int64, privateKey string) (*github.Client, string, error) {
	tr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, []byte(privateKey))
	if err != nil {
		return nil, "", errors.Wrap(err, "new transport")
	}

	client := github.NewClient(
		&http.Client{
			Transport: tr,
		},
	)

	token, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return nil, "", errors.Wrap(err, "create installation access token")
	}
	if token.Token == nil || *token.Token == "" {
		return nil, "", errors.New("empty token returned")
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
	return client, *token.Token, nil
}

type actionHandler func(ctx context.Context, config *conf.Config, payload *github.PullRequestEvent, client *github.Client, token string) error

func reportCommitStatus(ctx context.Context, config *conf.Config, payload *github.PullRequestEvent, handler actionHandler) {
	client, token, err := newGitHubClient(ctx, config.GitHubApp.AppID, *payload.Installation.ID, config.GitHubApp.PrivateKey)
	if err != nil {
		log.Error("Failed to create GitHub client: %v", err)
		return
	}

	createStatus := func(state, description string) {
		_, _, err = client.Repositories.CreateStatus(
			ctx,
			*payload.Repo.Owner.Login,
			*payload.Repo.Name,
			*payload.PullRequest.Head.SHA,
			&github.RepoStatus{
				State:       github.String(state),
				Description: github.String(description),
				Context:     github.String("Codenotify.run"),
			},
		)
		if err != nil {
			log.Error("Failed to create commit status on pull request %s: %v", *payload.PullRequest.HTMLURL, err)
			return
		}
	}

	createStatus("pending", "Running Codenotify")
	err = handler(ctx, config, payload, client, token)
	if err != nil {
		createStatus("error", err.Error())
		log.Error("Failed to run handler for pull request %s: %v", *payload.PullRequest.HTMLURL, err)
		return
	}
	createStatus("success", "Codenotify ran successfully")
}

func checkoutAndRun(ctx context.Context, config *conf.Config, payload *github.PullRequestEvent, token string) (string, error) {
	tmpPath := path.Join(os.TempDir(), fmt.Sprintf("codenotify.run-%s-%d", *payload.PullRequest.NodeID, time.Now().Unix()))
	err := os.MkdirAll(path.Dir(tmpPath), os.ModeDir)
	if err != nil {
		return "", errors.Wrap(err, "create temp directory")
	}
	defer func() { _ = os.RemoveAll(tmpPath) }()

	cloneURL, err := url.Parse(*payload.Repo.CloneURL)
	if err != nil {
		return "", errors.Wrap(err, "parse clone URL")
	}
	cloneURL.User = url.UserPassword("x-access-token", token)

	err = checkout(ctx, os.Stdout, tmpPath, cloneURL.String(), *payload.PullRequest.Head.SHA, *payload.PullRequest.Commits)
	if err != nil {
		return "", errors.Wrap(err, "checkout pull request")
	}

	output, err := codenotify(ctx, os.Stdout, config.Codenotify.BinPath, tmpPath, *payload.PullRequest.Base.SHA, *payload.PullRequest.Head.SHA)
	if err != nil {
		return "", errors.Wrap(err, "run Codenotify")
	}
	return output, nil
}

func handlePullRequestOpen(ctx context.Context, config *conf.Config, payload *github.PullRequestEvent, client *github.Client, token string) error {
	output, err := checkoutAndRun(ctx, config, payload, token)
	if err != nil {
		return errors.Wrap(err, "checkout and run")
	}

	if strings.Contains(output, "No notifications.") {
		return nil
	}

	comment, _, err := client.Issues.CreateComment(
		ctx,
		*payload.Repo.Owner.Login,
		*payload.Repo.Name,
		*payload.PullRequest.Number,
		&github.IssueComment{
			Body: github.String(output),
		},
	)
	if err != nil {
		return errors.Wrap(err, "create comment")
	}

	log.Info("Created comment %s", *comment.HTMLURL)
	return nil
}

func handlePullRequestSynchronize(ctx context.Context, config *conf.Config, payload *github.PullRequestEvent, client *github.Client, token string) error {
	output, err := checkoutAndRun(ctx, config, payload, token)
	if err != nil {
		return errors.Wrap(err, "checkout and run")
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
		return errors.Wrap(err, "list comments")
	}

	for _, comment := range comments {
		if comment.Body == nil || !strings.Contains(*comment.Body, `<!-- codenotify:CODENOTIFY report -->`) {
			continue
		}

		_, _, err = client.Issues.EditComment(
			ctx,
			*payload.Repo.Owner.Login,
			*payload.Repo.Name,
			*comment.ID,
			&github.IssueComment{
				Body: github.String(output),
			},
		)
		if err != nil {
			return errors.Wrap(err, "edit comment")
		}
		log.Info("Edited comment %s", *comment.HTMLURL)
		return nil
	}

	if strings.Contains(output, "No notifications.") {
		return nil
	}

	comment, _, err := client.Issues.CreateComment(
		ctx,
		*payload.Repo.Owner.Login,
		*payload.Repo.Name,
		*payload.PullRequest.Number,
		&github.IssueComment{
			Body: github.String(output),
		},
	)
	if err != nil {
		return errors.Wrap(err, "create comment")
	}

	log.Info("Created comment %s", *comment.HTMLURL)
	return nil
}

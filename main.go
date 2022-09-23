// Copyright 2022 Unknwon. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

		switch event {
		case "pull_request":
			var payload struct {
				Action      string `json:"action"`
				PullRequest struct {
					URL    string `json:"url"`
					Number int    `json:"number"`
				} `json:"pull_request"`
				Repository struct {
					Name  string `json:"name"`
					Owner struct {
						Login string `json:"login"`
					} `json:"owner"`
				} `json:"repository"`
				Installation struct {
					ID int64 `json:"id"`
				} `json:"installation"`
			}
			err = json.NewDecoder(r.Body).Decode(&payload)
			if err != nil {
				return http.StatusBadRequest, fmt.Sprintf("Failed to decode payload: %v", err)
			}
			if payload.Action != "opened" {
				return http.StatusOK, fmt.Sprintf("Event %q with action %q has been received but nothing to do", event, payload.Action)
			}

			if payload.Installation.ID <= 0 {
				return http.StatusBadRequest, "No installation ID"
			}

			go func() {
				err = createPullRequestComment(
					config.GitHubApp.AppID,
					payload.Installation.ID,
					config.GitHubApp.PrivateKey,
					payload.Repository.Owner.Login,
					payload.Repository.Name,
					payload.PullRequest.Number,
				)
				if err != nil {
					log.Error("Failed to create comment on pull request %s: %v", payload.PullRequest.URL, err)
					return
				}

				log.Info("Created comment on pull request %s", payload.PullRequest.URL)
			}()
			return http.StatusAccepted, http.StatusText(http.StatusAccepted)
		}
		return http.StatusOK, fmt.Sprintf("Event %q has been received but nothing to do", event)
	})
	f.Run()
}

func createPullRequestComment(appID, installationID int64, privateKey, owner, repo string, number int) error {
	tr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, []byte(privateKey))
	if err != nil {
		return errors.Wrap(err, "new transport")
	}

	client := github.NewClient(
		&http.Client{
			Transport: tr,
		},
	)

	ctx := context.Background()
	token, _, err := client.Apps.CreateInstallationToken(ctx, installationID, nil)
	if err != nil {
		return errors.Wrap(err, "create installation access token")
	}
	if token.Token == nil || *token.Token == "" {
		return errors.New("empty token returned")
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

	_, _, err = client.Issues.CreateComment(
		ctx,
		owner,
		repo,
		number,
		&github.IssueComment{
			Body: github.String("Hello world!"),
		},
	)
	if err != nil {
		return errors.Wrap(err, "create comment")
	}
	return nil
}

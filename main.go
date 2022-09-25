// Copyright 2022 Unknwon. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flamego/flamego"
	"github.com/google/go-github/v45/github"
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
			go handlePullRequestOpen(context.Background(), config, &payload)
		case "synchronize":
			go handlePullRequestSynchronize(context.Background(), config, &payload)
		default:
			return http.StatusOK, fmt.Sprintf("Event %q with action %q has been received but nothing to do", event, *payload.Action)
		}
		return http.StatusAccepted, http.StatusText(http.StatusAccepted)
	})
	f.Run()
}

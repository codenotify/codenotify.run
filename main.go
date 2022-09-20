// Copyright 2022 Unknwon. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"net/http"

	"github.com/flamego/flamego"
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

	f := flamego.Classic()
	f.Post("/-/webhook", func(r *http.Request) string {
		event := r.Header.Get("X-GitHub-Event")
		log.Trace("Received event: %s", event)
		return event
	})
	f.Run()
}

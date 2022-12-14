// Copyright 2022 Unknwon. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package conf

import (
	"strings"

	"github.com/go-ini/ini"
	"github.com/pkg/errors"

	"github.com/codenotify/codenotify.run/conf"
)

// Build time and commit information.
//
// ⚠️ WARNING: should only be set by "-ldflags".
var (
	BuildTime   string
	BuildCommit string
)

// Config contains all the configuration.
type Config struct {
	// Server contains the server configuration.
	Server struct {
		ExternalURL string `ini:"EXTERNAL_URL"`
		LogsRootDir string
	}
	// GitHubApp contains the GitHub App configuration.
	GitHubApp struct {
		AppID         int64  `ini:"APP_ID"`
		ClientID      string `ini:"CLIENT_ID"`
		ClientSecret  string
		PrivateKey    string
		WebhookSecret string
	}
	// Codenotify contains the Codenotify configuration.
	Codenotify struct {
		BinPath string
	}
}

// Load loads configuration from file.
func Load() (*Config, error) {
	data, err := conf.Files.ReadFile("app.ini")
	if err != nil {
		return nil, errors.Wrap(err, `read default "app.ini"`)
	}

	file, err := ini.LoadSources(
		ini.LoadOptions{
			IgnoreInlineComment: true,
		},
		data,
		"custom/conf/app.ini",
	)
	if err != nil {
		return nil, errors.Wrap(err, `load sources`)
	}
	file.NameMapper = ini.SnackCase

	var config Config
	if err = file.Section("server").MapTo(&config.Server); err != nil {
		return nil, errors.Wrap(err, `mapping "[server]" section`)
	} else if err = file.Section("github_app").MapTo(&config.GitHubApp); err != nil {
		return nil, errors.Wrap(err, `mapping "[github_app]" section`)
	} else if err = file.Section("codenotify").MapTo(&config.Codenotify); err != nil {
		return nil, errors.Wrap(err, `mapping "[codenotify]" section`)
	}

	config.Server.ExternalURL = strings.TrimSuffix(config.Server.ExternalURL, "/")
	return &config, nil
}

// Copyright 2022 Unknwon. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package conf

import (
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

// Config contains the configuration information.
type Config struct {
	GitHubApp struct {
		AppID        int64  `ini:"APP_ID"`
		ClientID     string `ini:"CLIENT_ID"`
		ClientSecret string
		PrivateKey   string
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
	if err = file.Section("github_app").MapTo(&config.GitHubApp); err != nil {
		return nil, errors.Wrap(err, `mapping "[github_app]" section`)
	}
	return &config, nil
}

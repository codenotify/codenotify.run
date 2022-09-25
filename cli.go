// Copyright 2022 Unknwon. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func run(ctx context.Context, w io.Writer, command string, args ...string) ([]byte, error) {
	cmdWithArgs := strings.Join(append([]string{command}, args...), " ")

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "running command %q", cmdWithArgs)
	}
	_, _ = fmt.Fprintln(w, cmdWithArgs)
	_, _ = fmt.Fprintln(w, string(out))
	return out, nil
}

func checkout(ctx context.Context, w io.Writer, repoPath, remoteURL, headCommit string, commitsCount int) error {
	out, err := run(ctx, w, "git", "init", repoPath)
	if err != nil {
		return fmt.Errorf("init: %v - %s", err, out)
	}

	_, err = run(ctx, w, "git", "-C", repoPath, "remote", "add", "origin", remoteURL)
	if err != nil {
		return fmt.Errorf("add remote: %v - %s", err, out)
	}

	_, err = run(
		ctx,
		w,
		"git",
		"-C", repoPath,
		"-c", "protocol.version=2",
		"fetch", "--no-tags", "--prune", "--no-recurse-submodules", "--quiet",
		"--deepen="+strconv.Itoa(commitsCount),
		"origin", headCommit,
	)
	if err != nil {
		return fmt.Errorf("fetch: %v - %s", err, out)
	}
	return nil
}

func codenotify(ctx context.Context, w io.Writer, binPath, repoPath, baseRef, headRef string) (string, error) {
	output, err := run(
		ctx,
		w,
		binPath,
		"--cwd", repoPath,
		"--baseRef", baseRef,
		"--headRef", headRef,
		"--format=markdown",
		"--filename=CODENOTIFY",
		"--subscriber-threshold=10",
		"--verbose",
	)
	if err != nil {
		return "", errors.Wrap(err, "run")
	}
	return string(output), nil
}

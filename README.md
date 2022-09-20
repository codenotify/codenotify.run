# codenotify.run

[Codenotify](https://github.com/sourcegraph/codenotify) as a Service.

## What?

Codenotify.run is a GitHub App backend of [Codenotify](https://github.com/sourcegraph/codenotify) that lets you subscribe to file changes in pull requests. It's a great way to stay up to date with changes to files you care about.

## Why?

The GitHub Action offered by the upstream [Codenotify](https://github.com/sourcegraph/codenotify) uses the designated `GITHUB_TOKEN` which has some shortcomings:

1. It cannot mention teams in pull request comments
2. It cannot post comments if the pull request is coming from a fork repository

Using a personal access token would of course solve the first problem, but a personal access token is too powerful and it is impractical to add it to all fork repositories (i.e. can't solve the second problem).

## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.

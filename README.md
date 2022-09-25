# codenotify.run

[Codenotify](https://github.com/sourcegraph/codenotify) as a Service.

## What?

Codenotify.run is a GitHub App backend of [Codenotify](https://github.com/sourcegraph/codenotify) that lets you subscribe to file changes in pull requests. It's a great way to stay up to date with changes to files you care about.

## Why?

The GitHub Action offered by the upstream [Codenotify](https://github.com/sourcegraph/codenotify) uses the designated `GITHUB_TOKEN` which has some shortcomings:

1. It cannot mention teams in pull request comments
2. It cannot post comments if the pull request is coming from a fork repository

Using a personal access token would of course solve the first problem, but a personal access token is too powerful and it is impractical to add it to all fork repositories (i.e. can't solve the second problem).

## How?

Install the [Codenotify](https://github.com/apps/codenotify) GitHub App on your repositories and add some [CODENOTIFY files](https://github.com/sourcegraph/codenotify#codenotify-files).

## Local development

### Step 1: Install dependencies

- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) (v2.3 or higher)
- [Go](https://golang.org/doc/install) (v1.19 or higher)
- [Task](https://github.com/go-task/task) (v3)
- [ngrok](https://ngrok.com/)
- [Codenotify](https://github.com/sourcegraph/codenotify) (v0.6.3 or higher)

#### macOS

1. Install [Homebrew](https://brew.sh/).
2. Install dependencies:

	```bash
	brew install git go go-task/tap/go-task ngrok
	go install https://github.com/sourcegraph/codenotify@latest
	
	# In the root directory of the repository
	ln -s $(go env GOPATH)/bin/codenotify $(pwd)/.bin/codenotify
	```

### Step 2: Create a test GitHub App

You need to use the `ngrok` to get a public URL for your local development server to be able to receive GitHub webhooks:

```bash
$ ngrok http 2830
```

Follow this [magic link](https://github.com/settings/apps/new?name=codenotify-test&url=https://codenotify.run&webhook_active=true&webhook_url=https://%3Cyour%20ngrok%20domain%3E/-/webhook&statuses=write&contents=read&pull_requests=write&emails=read&events[]=pull_request) to create your test GitHub App.

### Step 3: Start the server

```bash
$ task
```

By default, the server will listen on address `0.0.0.0:2830`, but you can change it by setting the `FLAMEGO_ADDR` environment variable:

```bash
$ FLAMEGO_ADDR=localhost:8888 task
```

Don't forget to restart your `ngrok` if you changed the port!

## Credits

The logo is a remix of [two](https://www.flaticon.com/free-icon/settings_5305761) [icons](https://www.flaticon.com/free-icon/notification_4270302) from [flaticon.com](https://www.flaticon.com/).

## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.

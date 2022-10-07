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

### Free public server

Free public server comes with absolutely shitty availability.

1. Install the [Codenotify](https://github.com/apps/codenotify) GitHub App on your repositories.
2. Add some [CODENOTIFY files](https://github.com/sourcegraph/codenotify#codenotify-files).

### Run your own server

Docker images for the Codenotify.run server are available both on [Docker Hub](https://hub.docker.com/r/unknwon/codenotify.run) and [GitHub Container Registry](https://github.com/codenotify/codenotify.run/pkgs/container/codenotify.run).

> **Note**
> The `latest` tag represents the latest build from the `main` branch.

You need to create a `custom` directory for the configuration file `app.ini`:

```bash
$ mkdir -p custom/conf
$ touch custom/conf/app.ini
```

Please refer to [Local development > Step 2: Create a test GitHub App](#step-2-create-a-test-github-app) for creating a GitHub App, setting up a reverse proxy and filling out necessary configuration options. View [`conf/app.ini`](conf/app.ini) for all available configuration options.

> **Note**
> The [Caddy web server](https://caddyserver.com/) is recommended for production use with automatic HTTPS.

Then volume the `custom` directory into the Docker container for it being able to start (`/app/codenotify.run/custom` is the path inside the container):

```bash
$ docker run \
    --name=codenotify.run \
    -p 12830:2830 \
    -v $(pwd)/custom:/app/codenotify.run/custom \
    unknwon/codenotify.run
```

## Local development

### Step 1: Install dependencies

- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) (v2.3 or higher)
- [Go](https://golang.org/doc/install) (v1.19 or higher)
- [Task](https://github.com/go-task/task) (v3)
- [ngrok](https://ngrok.com/)
- [Codenotify](https://github.com/sourcegraph/codenotify) (v0.6.4 or higher)

#### macOS

1. Install [Homebrew](https://brew.sh/).
2. Install dependencies:

	```bash
	brew install git go go-task/tap/go-task ngrok
	go install https://github.com/sourcegraph/codenotify@v0.6.4

	# In the root directory of the repository
	ln -s $(go env GOPATH)/bin/codenotify $(pwd)/.bin/codenotify
	```

### Step 2: Create a test GitHub App

You need to use the `ngrok` to get a public URL for your local development server to be able to receive GitHub webhooks:

```bash
$ ngrok http 2830
```

Follow this [magic link](https://github.com/settings/apps/new?name=codenotify-test&url=https://codenotify.run&webhook_active=true&webhook_url=https://%3Cyour%20ngrok%20domain%3E/-/webhook&statuses=write&contents=read&pull_requests=write&emails=read&events[]=pull_request) to create your test GitHub App.

Once you have created your test GitHub App, put the **App ID** and [**Private key**](https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#generating-a-private-key) in the `custom/conf/app.ini` file:

```ini
[github_app]
APP_ID = 123456
PRIVATE_KEY = """-----BEGIN RSA PRIVATE KEY-----
...
-----END RSA PRIVATE KEY-----"""
```

### Step 3: Start the server

```bash
$ task
```

By default, the server will listen on address `0.0.0.0:2830`, but you can change it by setting the `FLAMEGO_ADDR` environment variable:

```bash
$ FLAMEGO_ADDR=localhost:8888 task
```

Don't forget to restart your `ngrok` and update your `custom/conf/app.ini` file before you changed the port!

```ini
[server]
EXTERNAL_URL = http://localhost:8888
```

## Credits

The logo is a remix of [two](https://www.flaticon.com/free-icon/settings_5305761) [icons](https://www.flaticon.com/free-icon/notification_4270302) from [flaticon.com](https://www.flaticon.com/).

## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.

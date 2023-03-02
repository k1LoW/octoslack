# octoslack

`octoslack` is a tool for transforming HTTP requests from any webhook into Slack messages.

## Usage

``` mermaid
flowchart TB
    src[Some webhook source] -- POST https://octoslack.example.com/services/XXX/YYY --- payload[JSON payload]
    subgraph "octoslack.example.com"
    payload[JSON payload] -- "Transform payload by octoslack" --- spayload[JSON payload for Slack]
    end
    spayload[JSON payload for Slack] -- POST https://hooks.slack.com/services/XXX/YYY --> Slack[Slack Incoming Webhook endpoint]
```

### 0. Requirements

- Slack Incoming Webhook URL ( `https://hooks.slack..com/services/XXX/YYY` )

### 1. Setup config.yml

Describe the settings for converting HTTP requests from the target webhook.

In particular, octoslack targets [GitHub repository webhooks](https://docs.github.com/en/rest/webhooks?apiVersion=2022-11-28), so it parses the `X-GitHub-Event` header ( to `github_event` ).

``` yaml
# config.yml
requests:
  -
    condition: github_event == 'discussion' && payload.action == 'created'
    transform:
      blocks:
        - type: section
          text:
            type: mrkdwn
            text: |-
              Discussion created by {{ payload.user.login }}
        - type: section
          text:
            type: mrkdwn
            text: |-
              {{ quote(payload.discussion.body) }}
```

### 2. Start octoslack server

Start the octoslack server and make the server accessible from the Internet.

``` console
$ octoslack server -c config.yml
```

If you want to use a Docker image and start the server using the config file in the GitHub repository, you can run the following.

``` sh
$ docker container run -it --rm --name octoslack-server \
  -e OCTOSLACK_CONFIG=github://k1LoW/octoslack/testdata/config.yml \
  -e OCTOSLACK_PORT=8080 \
  -e GITHUB_TOKEN \ # use GITHUB_TOKEN for getting config.yml from GitHub repository
  -p 8080:8080 \
  ghcr.io/k1low/octoslack:latest
```

Here, assume it is published as `https://octoslack.example.com`.

### 3. Set Slack Incoming webhook URL by changing the hostname to `octoslack.example.com`

Change the hostname of the Slack Incoming webhook URL to `octoslack.example.com`.

`https://hooks.slack..com/services/XXX/YYY` -> `https://octoslack.example.com/services/XXX/YYY`

And set it as the destination URL.

### 4. Webhook event fired

HTTP requests are transformed into requests that Slack can read through octoslack.

``` mermaid
flowchart TB
    src[GitHub Webhooks] -- POST https://octoslack.example.com/services/XXX/YYY --- payload[JSON payload]
    subgraph "octoslack.example.com"
    payload[JSON payload] -- "Transform payload by octoslack" --- spayload[JSON payload for Slack]
    end
    spayload[JSON payload for Slack] -- POST https://hooks.slack.com/services/XXX/YYY --> Slack[Slack Incoming Webhook endpoint]
```

# Harbor-Telegram-Bot

[![CI](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/ci.yml/badge.svg)](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/ci.yml)
[![Lint](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/golangci-lint.yml)
[![CodeQL](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/github-code-scanning/codeql)
[![Released](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/release.yml/badge.svg)](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/release.yml)

Harbor event notifications for Telegram.

## Release 3.1.0

- Filter SHA digest tag notifications — Harbor 2.8+ fires an interim webhook with a `sha256:...` tag when pushing OCI Helm artifacts before renaming to the real tag. Bot now silently drops these to prevent duplicate notifications.
- Updated Go runtime to 1.26.

## Release 3.0.0

- Support **`Artifact pushed`** option - [PUSH_ARTIFACT](https://goharbor.io/docs/2.13.0/working-with-projects/project-configuration/configure-webhooks/#:~:text=artifact%20to%20registry-,PUSH_ARTIFACT,-Repository%20namespace%20name) event type.
- Support **`Artifact pulled`** option - [PULL_ARTIFACT](https://goharbor.io/docs/2.13.0/working-with-projects/project-configuration/configure-webhooks/#:~:text=artifact%20from%20registry-,PULL_ARTIFACT,-Repository%20namespace%20name) event type.
- Support **`Artifact deleted`** option - [DELETE_ARTIFACT](https://goharbor.io/docs/2.13.0/working-with-projects/project-configuration/configure-webhooks/#:~:text=artifact%20from%20registry-,DELETE_ARTIFACT,-Repository%20namespace%20name) event type.
- Support **`Quota exceed`** option - [QUOTA_EXCEED](https://goharbor.io/docs/2.13.0/working-with-projects/project-configuration/configure-webhooks/#:~:text=Project%20quota%20exceeded-,QUOTA_EXCEED,-Repository%20namespace%20name) event type.
- Support **`Quota near threshold`** option - [QUOTA_WARNING](https://goharbor.io/docs/2.13.0/working-with-projects/project-configuration/configure-webhooks/#:~:text=quota%20near%20threshold-,QUOTA_WARNING,-Repository%20namespace%20name) event type.

## Getting started

### Pre-requirements

- Harbor release v2.10+ (v2.12+ tested)
- Harbor ChartMuseum Extension
- Create Telegram Bot with [BotFather](https://core.telegram.org/bots/features#botfather)
- Get Bot API Token
- Get your ChatID (example):
  - public: `https:/t.me/MY_CHAT`
  - [private](https://telegram-bot-sdk.readme.io/reference/getupdates): `-2233445566778`
- Add your bot to chanel or group with admin rules (messages access).

#### Optional

- TopidID ([message_thread_id](https://core.telegram.org/bots/api#message)):

  - Default or General Topic: `0`

  - To test with topics:

        ```bash:
        curl -X GET 'https://api.telegram.org/bot<bot-api-token>/sendMessage?chat_id=<chat-id>&message_thread_id=<chat-topic-id>&text=HelloTopic!'
        ```

### Install

1. To start Bot, run the following command with your variables in terminal:

    ``` bash
    docker run -it -p 441:441
        --name harbor-telegram-bot
        -e CHAT_ID=<chat-id>
        -e BOT_TOKEN=<bot-api-token>
        -e TOPIC_ID=<topic-id>
        -e HARBOR_URL=<harbor-url> # https://hub.harbor.com
        -e HARBOR_USER=<harbor-user>
        -e HARBOR_PASS=<harbor-pass>
        -v /<certs-path>:/usr/local/share/ca-certificates #for custom ca certificates
        alexpokatilov/harbor-telegram-bot:3.1.0
    ```
    Set `-e WARN_ON_PUSH=true`, if you want to see usage quota warning with push event.
    Set `-e DEBUG=true`, if you want to see all logs with raw format.
    Set `-e FILTER_SHA_TAGS=false`, to disable filtering of SHA digest tag notifications (enabled by default).

2. Configure your Harbor `http` webhook

3. Check your bot. Send [POST](#json-payload-format) request to `http://<hostname>:441/webhook-bot`
4. Bot message example:

    - Docker Image

        ```text
        🐳 New image pushed by: admin
        • Host: hub.harbor.com
        • Project: test-webhook
        • Repository: test-webhook/debian
        • Tag: latest
        ```

        ```text
        🐳 New image pushed by: admin
        • Host: hub.harbor.com
        • Project: test-webhook
        • Repository: test-webhook/debian
        • Tag: latest

        ⚠️ Warning!! Quota usage reach 85%!!
        • Details: quota usage reach 95.19%: resource storage used 47.60 MB of 50.00 MB
        ```

        ```text
        🐳 Artifact pulled by: admin
        • Host: hub.harbor.com
        • Access: public
        • Project: test-webhook
        • Access: public
        • Repository: test-webhook/debian
        • Tag: latest
        ```

        ```text
        ❗️ Attention!
        📦 Artifact removed by: admin
        • Access: public
        • Host: hub.harbor.com
        • Project: test-webhook
        • Repository: test-webhook/debian
        • Tag: latest
        ```

    - Helm Chart

        ```text
        ☸️ New chart version uploaded by: admin
        • Host: hub.harbor.com
        • Access: public
        • Project: test-webhook
        • Repository: test-webhook/debian
        • Version: 0.1.0
        ```

        ```text
        ☸️ Chart pulled by: admin
        • Host: hub.harbor.com
        • Access: public
        • Project: test-webhook
        • Repository: test-webhook/debian
        • Version: 0.1.0
        ```

        ```text
        ❗️ Attention!
        ☸️ Chart removed by: admin
        • Host: hub.harbor.com
        • Access: public
        • Project: test-webhook
        • Repository: test-webhook/debian
        • Version: 0.1.0
        ```

    - Alert (QUOTA)

        ```text
        🚨 Alert!!! Project quota has been exceeded!!!
        • Project: test-webhook
        • Details: adding 30.1 MiB of storage resource, which when updated to current usage of 1 GiB will exceed the configured upper limit of 1 GiB.
        ```

        ```text
        ⚠️ Warning!! Quota usage reach 85%!!
        • Project: test-webhook
        • Details: quota usage reach 85%: resource storage used 0.9 GiB of 1 GiB
        ```

## Links

- [Releases](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/releases)
- [Docker Hub](https://hub.docker.com/r/alexpokatilov/harbor-telegram-bot)

### Ref links

- [Harbor - goharbor.io/docs/2.13.0/working-with-projects/project-configuration/configure-webhooks/](https://goharbor.io/docs/2.13.0/working-with-projects/project-configuration/configure-webhooks/)

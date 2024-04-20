# Harbor-Telegram-Bot

[![Build](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/docker.yml/badge.svg)](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/docker.yml)
[![Lint](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/golangci-lint.yml)
[![CodeQL](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/github-code-scanning/codeql)
[![Released](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/release.yml/badge.svg)](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/actions/workflows/release.yml)

Harbor event notifications for Telegram.

## Release 2.3.0

- Support **`Artifact pushed`** option - [PUSH_ARTIFACT](https://goharbor.io/docs/2.7.0/working-with-projects/project-configuration/configure-webhooks/#:~:text=artifact%20to%20registry-,PUSH_ARTIFACT,-Repository%20namespace%20name) event type.
- Support **`Artifact pulled`** option - [PULL_ARTIFACT](https://goharbor.io/docs/2.7.0/working-with-projects/project-configuration/configure-webhooks/#:~:text=artifact%20from%20registry-,PULL_ARTIFACT,-Repository%20namespace%20name) event type.
- Support **`Artifact deleted`** option - [DELETE_ARTIFACT](https://goharbor.io/docs/2.7.0/working-with-projects/project-configuration/configure-webhooks/#:~:text=artifact%20from%20registry-,DELETE_ARTIFACT,-Repository%20namespace%20name) event type.
- Support **`Chart uploaded`** option - [UPLOAD_CHART](https://goharbor.io/docs/2.7.0/working-with-projects/project-configuration/configure-webhooks/#:~:text=chart%20to%20chartMuseum-,UPLOAD_CHART,-Repository%20name%2C%20chart) event type.
- Support **`Chart downloaded`** option - [DOWNLOAD_CHART](https://goharbor.io/docs/2.7.0/working-with-projects/project-configuration/configure-webhooks/#:~:text=chart%20from%20chartMuseum-,DOWNLOAD_CHART,-Repository%20name%2C%20chart) event type.
- Support **`Chart deleted`** option - [DELETE_CHART](https://goharbor.io/docs/2.7.0/working-with-projects/project-configuration/configure-webhooks/#:~:text=chart%20from%20chartMuseum-,DELETE_CHART,-Repository%20name%2C%20chart) event type.
- Support **`Quota exceed`** option - [QUOTA_EXCEED](https://goharbor.io/docs/2.7.0/working-with-projects/project-configuration/configure-webhooks/#:~:text=Project%20quota%20exceeded-,QUOTA_EXCEED,-Repository%20namespace%20name) event type.
- Support **`Quota near threshold`** option - [QUOTA_WARNING](https://goharbor.io/docs/2.7.0/working-with-projects/project-configuration/configure-webhooks/#:~:text=quota%20near%20threshold-,QUOTA_WARNING,-Repository%20namespace%20name) event type.

## Getting started

### Pre-requirements

- Harbor v2.0 - v2.7
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
        alexpokatilov/harbor-telegram-bot:2.3.0
    ```

    Set `-e DEBUG=true`, if you want to see all logs with raw format.

2. Configure your Harbor `http` webhook

3. Check your bot. Send [POST](#json-payload-format) request to `http://<hostname>:441/webhook-bot`
4. Bot message example:

    - Docker Image (ARTIFACT)

        ```text
        🐳 New image pushed by: admin
        • Host: hub.harbor.com
        • Project: test-webhook
        • Repository: test-webhook/debian
        • Tag: latest
        ```

        ```text
        🐳 Artifact pulled by: admin
        • Host: hub.harbor.com
        • Project: test-webhook
        • Access: public
        • Repository: test-webhook/debian
        • Tag: latest
        ```

        ```text
        ❗️ Attention!
        🐳 Artifact removed by: admin
        • Host: hub.harbor.com
        • Project: test-webhook
        • Repository: test-webhook/debian
        • Tag: latest
        ```

    - Helm Chart (CHART)

        ```text
        ☸️ New chart version uploaded by: admin
        • Host: hub.harbor.com
        • Project: test-webhook
        • Chart: test-webhook/debian
        • Version: latest
        ```

        ```text
        ☸️ Chart pulled by: admin
        • Host: hub.harbor.com
        • Project: test-webhook
        • Access: public
        • Chart: debian
        • Version: 0.1.0
        ```

        ```text
        ❗️ Attention!
        ☸️ Chart removed by: admin
        • Host: hub.harbor.com
        • Project: test-webhook
        • Chart: test-webhook/debian
        • Version: latest
        ```

    - Alert (QUOTA)

        ```text
        🚨 Alert!!! Project quota has been exceed!!!
        • Project: test-webhook
        ```

        ```text
        ⚠️ Warning!! Quota usage reach 85%!!
        • Project: test-webhook
        ```

## Links

- [Releases](https://github.com/AlexPokatilov/Harbor-Telegram-Bot/releases)
- [Docker Hub](https://hub.docker.com/r/alexpokatilov/harbor-telegram-bot)

### Development

#### Json Payload Format

- [Artifact deleted](./readme/PayloadFormat/DELETE_ARTIFACT.json)
- [Artifact pulled](./readme/PayloadFormat/PULL_ARTIFACT.json)
- [Artifact pushed](./readme/PayloadFormat/PULL_ARTIFACT.json)
- [Chart deleted](./readme/PayloadFormat/DELETE_CHART.json)
- [Chart downloaded](./readme/PayloadFormat/DOWNLOAD_CHART.json)
- [Chart uploaded](./readme/PayloadFormat/UPLOAD_CHART.json)
- [Quota exceed](./readme/PayloadFormat/QUOTA_EXCEED.json)
- [Quota near threshold](./readme/PayloadFormat/QUOTA_WARNING.json)
- [Scanning failed](./readme/PayloadFormat/SCANNING_FAILED.json)
- [Scanning finished](./readme/PayloadFormat/SCANNING_COMPLETED.json)
- [Scanning stopped](./readme/PayloadFormat/SCANNING_STOPPED.json)
- [Replication finished](./readme/PayloadFormat/REPLICATION.json)
- [Tag retention finished](./readme/PayloadFormat/TAG_RETENTION_FINISHED.json)

### Ref links

- [github.com/go-telegram-bot-api/telegram-bot-api/v5](https://pkg.go.dev/github.com/go-telegram-bot-api/telegram-bot-api/v5@v5.5.1)
- [github.com/technoweenie/multipartstreamer](https://pkg.go.dev/github.com/technoweenie/multipartstreamer@v1.0.1)
- [Harbor - goharbor.io/docs/2.7.0/working-with-projects/project-configuration/configure-webhooks/](https://goharbor.io/docs/2.7.0/working-with-projects/project-configuration/configure-webhooks/)

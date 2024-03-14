# :airplane: Telegram Butler

[![CI Status](https://github.com/GolangUA/telegram-butler/actions/workflows/ci.yml/badge.svg)](https://github.com/GolangUA/telegram-butler/actions/workflows/ci.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=GolangUA_telegram-butler&metric=alert_status)](https://sonarcloud.io/dashboard?id=GolangUA_telegram-butler)
[![Go Report](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)](https://goreportcard.com/report/github.com/GolangUA/telegram-butler)
<br>
[![Go Version](https://img.shields.io/github/go-mod/go-version/GolangUA/telegram-butler?logo=go)](go.mod)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=GolangUA_telegram-butler&metric=coverage)](https://sonarcloud.io/dashboard?id=GolangUA_telegram-butler)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=GolangUA_telegram-butler&metric=code_smells)](https://sonarcloud.io/dashboard?id=GolangUA_telegram-butler)
[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=GolangUA_telegram-butler&metric=ncloc)](https://sonarcloud.io/dashboard?id=GolangUA_telegram-butler)

Telegram bot for managing GolangUA community

## Local development

1. Update the values of all needed environment variables in `.env.local` file:

    ```dotenv
    NGROK_AUTHTOKEN=ngrok-token
    BOT_TOKEN=telegram-bot-token
    ```
   > To get `NGROK_AUTHTOKEN` visit: https://dashboard.ngrok.com/get-started/your-authtoken

2. Run

   ```shell
   go run -tags local ./cmd/telegram-butler/
   ```

   Or

   ```shell
   task run:local
   ```

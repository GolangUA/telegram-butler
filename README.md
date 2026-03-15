# :airplane: Telegram Butler

[![CI Status](https://github.com/GolangUA/telegram-butler/actions/workflows/ci.yml/badge.svg)](https://github.com/GolangUA/telegram-butler/actions/workflows/ci.yml)
[![Go Report](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)](https://goreportcard.com/report/github.com/GolangUA/telegram-butler)
[![Go Version](https://img.shields.io/github/go-mod/go-version/GolangUA/telegram-butler?logo=go)](go.mod)

Telegram bot for managing GolangUA community

## Local development

1. Create new `.env` file and copy envs from `example.env` replacing with proper values

2. Run

   ```shell
   go run -tags local ./cmd/telegram-butler/
   ```

   Or

   ```shell
   task run:local
   ```

3. Run in Docker

   ```shell
   task run:docker
   ```

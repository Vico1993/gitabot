# Gitabot

It's a program made to fetch dependabot PR. Check if all check passed and approved or merged if selected

## Table of Contents

- [Getting Started](#getting-started)
- [Usage](#usage)
- [Contributing](#contributing)
- [Usage](#usage)

## Getting Started

To get started with gitabot, clone the repository to your local machine:

```sh
git clone https://github.com/Vico1993/gitabot.git
cd gitabot
```

## Prerequisites

Make sure you have the following tools installed on your machine:

- Go (at least version 1.21)
- Setup an .env file

```bash
# We need a Github token which is allow to read / write pull request
GITHUB_TOKEN=

# Bot can send some Telegram notification
# If wanted use
TELEGRAM_DISABLE=0
TELEGRAM_CHAT_ID=
TELEGRAM_BOT_TOKEN=
TELEGRAM_THREAT_ID=
```

## Installing

To install Gitabot, run the following command:

```sh
make ensure_deps
```

## Usage

To use Gitabot, run the following command:

```sh
make build && ./bin/bot
```

## Contributing

Contributions are welcome! Please see the [CONTRIBUTING.md](./CONTRIBUTING.md) file for more information.

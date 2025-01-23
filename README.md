# Gitabot

This script is designed to parse a list of GitHub issues and identify open pull requests created by Dependabot. Its primary function is to automatically approve these PRs if all associated checks have passed. If you are the owner of the repository, the script will merge them for you.

To merge pull requests, you need to be the owner of the repository and have AUTO_MERGE=1 set in your environment file.

While automating the approval and merging of PRs might seem risky, I'm a careful developer who trusts my CI pipeline. If the CI passes all checks, there's no reason for me to manually handle these PRs when a script can efficiently do it for me.

## Table of Contents

- [Setup Instructions](#setup-instructions)
- [Contributing](#contributing)

## Setup Instructions

Follow the steps below to set up and run the Automated Dependabot PR Approver:

### 1. Configure Environment Variables

- Create or update a `.env` file in the project root with the following variables:

  - **GitHub Credentials:**

    - `GITHUB_TOKEN`: Your GitHub API token.
    - `GITHUB_USERNAME`: Your GitHub username.

  - **Telegram Notifications:**

    - `TELEGRAM_CHAT_ID`: The chat ID where notifications will be sent.
    - `TELEGRAM_BOT_TOKEN`: The token for your Telegram bot.

### 2. Install Dependencies

- Run the following command to ensure all dependencies are installed:

```sh
  make ensure_deps
```

### 3. Build and Run the Script

- Build the project and run the bot with the following commands:

```sh
  make build && ./bin/bot
```

## Contributing

Contributions are welcome! Please see the [CONTRIBUTING.md](./CONTRIBUTING.md) file for more information.

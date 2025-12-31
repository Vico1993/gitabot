# Gitabot

Gitabot is an automated Go-based bot that finds, approves, and optionally merges Dependabot pull requests across your GitHub repositories. It intelligently processes PRs by checking CI status, mergeability, and existing approvals before taking action.

## Features

- 🔍 **Smart PR Discovery**: Automatically finds open Dependabot PRs in:
  - Repositories owned by your organization
  - Repositories where you're requested for review
- ✅ **Automatic Approval**: Approves PRs that meet all criteria:
  - All CI checks have passed
  - PR is mergeable (no conflicts)
  - Not already approved by you
- 🔀 **Auto-Merge Support**: Optionally merges approved PRs if you're the repository owner
- 🔄 **Rebase Management**: Automatically requests rebase for subsequent PRs after merging
- 📊 **Telegram Notifications**: Sends summary reports via Telegram (optional)
- ⚡ **Concurrent Processing**: Processes multiple repositories in parallel for efficiency

## How It Works

1. Searches GitHub for open Dependabot PRs you have access to
2. Groups PRs by repository for efficient processing
3. For each PR, verifies:
   - All CI checks are passing
   - PR is mergeable (no merge conflicts)
   - You haven't already approved this commit
4. Approves PRs that meet all criteria
5. If you're the repository owner and `AUTO_MERGE=1` is set, merges the first PR and requests rebase for others
6. Sends a summary report via Telegram (if configured)

## Table of Contents

- [Setup Instructions](#setup-instructions)
- [Environment Variables](#environment-variables)
- [Usage](#usage)
- [Development](#development)
- [Contributing](#contributing)

## Setup Instructions

### Prerequisites

- Go 1.21 or later
- A GitHub Personal Access Token with appropriate permissions
- (Optional) A Telegram bot token for notifications

### 1. Clone the Repository

```sh
git clone https://github.com/Vico1993/gitabot.git
cd gitabot
```

### 2. Install Dependencies

```sh
make ensure_deps
```

### 3. Configure Environment Variables

Create a `.env` file in the project root with the required variables (see [Environment Variables](#environment-variables) below).

### 4. Build and Run

```sh
make build && ./bin/bot
```

## Environment Variables

### Required

- `GITHUB_TOKEN`: Your GitHub Personal Access Token with `repo` scope
- `GITHUB_USERNAME`: Your GitHub username

### Optional

- `AUTO_MERGE`: Set to `"1"` to enable automatic merging of approved PRs (only works if you're the repository owner)
- `TELEGRAM_CHAT_ID`: Telegram chat ID where notifications will be sent
- `TELEGRAM_BOT_TOKEN`: Token for your Telegram bot
- `TELEGRAM_THREAT_ID`: (Note: typo in code) Telegram thread ID for threaded replies
- `TELEGRAM_DISABLE`: Set to `"1"` to disable Telegram notifications

## Usage

### Basic Usage

Simply run the bot:

```sh
./bin/bot
```

The bot will:

1. Search for all open Dependabot PRs
2. Process them according to the configured rules
3. Display a summary in the terminal
4. Send notifications via Telegram (if configured)

### Makefile Commands

- `make ensure_deps` - Install and vendor dependencies
- `make build` - Build the bot binary
- `make test` - Run tests
- `make lint` - Run linter
- `make lint_fix` - Run linter with automatic fixes
- `make watch` - Watch for changes and rebuild

## Development

### Running Tests

```sh
make test
```

### Linting

```sh
make lint
# or with auto-fix
make lint_fix
```

### Watch Mode

For development, use watch mode to automatically rebuild on changes:

```sh
make watch
```

## Contributing

Contributions are welcome! Please see the [CONTRIBUTING.md](./CONTRIBUTING.md) file for more information.

## Disclaimer

While automating the approval and merging of PRs might seem risky, this tool is designed for developers who trust their CI pipeline. If your CI passes all checks and the PR is mergeable, Gitabot can efficiently handle these routine tasks for you. Always review your CI configuration and ensure it provides adequate protection before enabling auto-merge.

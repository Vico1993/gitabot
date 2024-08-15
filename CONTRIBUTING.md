# Contributing to Otto

Thank you for your interest in contributing to _Otto_! Please take a moment to review these guidelines before submitting any contributions.

## Getting Started

To get started, fork the repository and clone your fork to your local machine.

```sh
git clone https://github.com/YOUR_USERNAME/YOUR_FORK.git
cd YOUR_FORK
```

## Making Changes

Before making any changes, please ensure that you are working on the latest version of the codebase by pulling from the upstream repository:

```sh
git remote add upstream https://github.com/Vico1993/otto-cron-feeds.git
git remote add upstream https://github.com/Vico1993/otto-cron-feeds.git
git fetch upstream
git checkout main
git merge upstream/main
```

Please make your changes on a new branch and ensure that your code adheres to our [code style guidelines](#code-style-guidelines):

```sh
git checkout -b my-feature-branch
```

Once you have made your changes, please submit a [pull request](https://github.com/Vico1993/otto-cron-feeds/pulls) with a clear description of the changes you have made and the rationale behind them.

## Code Style Guidelines

We adhere to the Go [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) for code style. Please ensure that your code adheres to these guidelines before submitting a pull request.

## Testing

Please ensure that your changes are accompanied by appropriate tests, and that all tests pass before submitting a pull request:

```sh
make test
```

## Issue Tracking

Please use the [issue tabs](https://github.com/Vico1993/otto-cron-feeds/issues) to report any bugs or issues you encounter. When reporting an issue, please provide a clear description of the problem, including steps to reproduce the issue.

## License

By contributing to [Otto](https://github.com/Vico1993/otto-cron-feeds), you agree that your contributions will be licensed under the [LICENSE](https://github.com/Vico1993/otto-cron-feeds/blob/main/LICENSE) file in the root directory of this repository.

We look forward to your contributions!

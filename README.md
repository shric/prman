2021-01-02: Status is pre-alpha. It works for me but was hacked up in an evening.

prman is an ugly but useful CLI tool. It lists all your PRs matching a search.

It shows the PR URL, the status, whether it has an approving reviewer, and
whether it's blocking on anything (e.g. CI/CD builds)

## Install

```shell
$ go get github.com/shric/prman/cmd/prman
```

## Run

```shell
$ prman 'author:shric is:open "some optional search string"'
```

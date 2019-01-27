# Beer Review

CLI for managing your JIRA / Gerrit / git workflow.

This is a Go port of [beer-review](https://github.com/kunickiaj/beer-review)

[![Build Status](https://gitlab.com/kunickiaj/beer/badges/master/pipeline.svg)](https://gitlab.com/kunickiaj/beer/commits/master)

[![Coverage Report](https://gitlab.com/kunickiaj/beer/badges/master/coverage.svg)](https://gitlab.com/kunickiaj/beer/commits/master)

## Prerequisites

Requires libgit2 v0.27 installed. On macOS `brew install libgit2`.

## Installation

`go get -u github.com/kunickiaj/beer`

## Configuration

By default, `beer` looks for a configuration file at `~/.beer.yaml` but an alternate path can also be specified with the `--config` flag.

The file is YAML containing two sections (mappings) with configuration for JIRA and Gerrit. An example is provided below.

```yaml
jira:
  url: https://issues.apache.org
  username: alice
  password: Password123!
gerrit:
  url: https://gerrit.googlesource.com
```

## Usage

All help is accessible by specifying the `--help` flag to any beer command/subcommand. `beer --help` will provide an overview of available commands.

### Common Workflow

#### Work on an Existing JIRA issue

`beer brew PRJ-1234` will create a new work branch from issue PRJ-1234 and insert an empty commit with the issue key followed by the issue summary as the commit message. It will also transition the JIRA issue to an In Progress state.

#### Work on a New JIRA issue

`beer brew -t Bug -s 'My issue summary' -d 'My detailed issue description` will create a new JIRA issue of type Bug, with the specified summary and detailed description. it will then create a new work branch from the newly created issue with the issue key followed by the issue summary as the commit message. It will also transition the JIRA issue to an In Progress state.

See the output of `beer brew --help` for all available flags.

#### Prepare for review

At this point you'll make your changes as usual before until you are ready to post a review. Your commits should be squashed and amend the empty commit that was automatically created.

For example: `git commit -a --amend`

### Create a new review

`beer taste` will push a review to the configured Gerrit server. The `--wip` flag is available if you wish to push a WIP review.

### Submit a change

TODO: Add `drink` command.
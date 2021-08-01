# Beer Review

CLI for managing your JIRA / Gerrit / git workflow.

This is a Go port of [beer-review](https://github.com/kunickiaj/beer-review)

![Release Status](https://github.com/kunickiaj/beer/actions/workflows/release.yaml/badge.svg)
![CI Status](https://github.com/kunickiaj/beer/actions/workflows/pr.yaml/badge.svg)
![CodeQL Status](https://github.com/kunickiaj/beer/actions/workflows/codeql-analysis.yml/badge.svg)

## Prerequisites

None

## Installation

`go get -u github.com/kunickiaj/beer`

or with Homebrew

`brew install kunickiaj/beer/beer`

## Configuration

By default, `beer` looks for a configuration file at `~/.beer.yaml` but an alternate path can also be specified with the `--config` flag.

The file is YAML containing two sections (mappings) with configuration for JIRA and Gerrit. An example is provided below.

```yaml
reviewTool: gerrit
jira:
  url: https://issues.apache.org/jira
  username: alice
gerrit:
  url: https://gerrit.googlesource.com # (optional, currently unused)
# optional section, you can specify persistent defaults for some flags
defaults:
  # beer will use 'trunk' for creating reviews instead of the default of 'main' 
  branch: trunk
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

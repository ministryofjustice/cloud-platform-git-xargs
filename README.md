[![Build status](https://github.com/ministryofjustice/cloud-platform-git-xargs/actions/workflows/release.yaml/badge.svg)](https://github.com/ministryofjustice/cloud-platform-git-xargs/actions/workflows/release.yaml)
[![tests](https://github.com/ministryofjustice/cloud-platform-git-xargs//workflows/tests/badge.svg)](https://github.com/ministryofjustice/cloud-platform-git-xargs/actions?query=workflow%3A%22tests%22)

## Problem statement

While trying to upgrade the terraform estate across the whole [cloud platform](https://github.com/ministryofjustice/cloud-platform) estate we noticed a number of repetitive steps that could be automated. We had to manually:

- Clone a repository
- Run `terraform 0.13upgrade`
- Stage, commit and push
- PR

We have two fundamental constraints when performing the above:

- We have over 50 repositories with the name `cloud-platform-terraform*` that contain terraform HCL.
- We maintain a multi-tenant Kubernetes cluster with both production and non-production code.

We need the ability to pass a pattern of a repository name, such as `*terraform`, as an argument to a cli with the ability to also skip and cherry pick commits to craft a pull request. As the Ministry of Justice run a multi-tenancy Kubernetes cluster with both production and non-production code, sometimes it's important to slice up your PR, only commiting certain changes that effect non-production namespaces and leave others. So having the autonomy to commit to my local machine and cherry pick which changes I need is essential.

### What tools are currently out there

There's a really magnificent repository called [git-xargs](https://github.com/gruntwork-io/git-xargs/). This tool performs the clone, run, push and pr abilities of the code in this repository. It doesn't however let us pass our repository name pattern and the skip feature doesn't allow you to go back and automate your PR creation later.

We also decided to use the [cobra](https://github.com/spf13/cobra) cli package as the team maintaining this repository will have experience using this tool. The maintainers of this repository will keep a close eye on the [git-xargs](https://github.com/gruntwork-io/git-xargs/) and will close our repository down if the feature set begins to align further.

## How to use it

You must have valid GitHub token. See the guide on [GitHub Personal access token](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token). For example:

```bash
export GITHUB_OAUTH_TOKEN=<your-personal-access-token>
```

- if you don't have the environment variable set, the app will tell you to set it.

Using the example in the problem statement, you'd run:

```bash
cloud-platform-git-xargs run -command "terraform 0.13upgrade" \
                             -organisation ministryofjustice \
                             -repository cloud-platform-terraform \
                             -loop \
                             -message "Upgrade Terraform HCL to Terraform 0.13.x"
```

This performs the following:

- Identify all repositories in the `ministryofjustice` organisation that contain the name `cloud-platform-terraform` (an example of this would be `cloud-platform-terraform-rds-instance`).
- Will clone each repository down to a temporary directory called `tmp/`.
- On each directory, run the `terraform 0.13upgrade` command.
- Commit using the message "Upgrade Terraform HCL to Terraform 0.13.x".
- Push a branch named `update` to the repository on GitHub.
- Create a PR.

### Flags to use

```bash
Flags:
  -c, --command string        the command you'd like to execute i.e. touch file
  -m, --commit string         the commit message you'd like to make (default "perform command on repository")
  -f, --file string           path to file containing list of repositories to process.
  -h, --help                  help for run
  -l, --loop-dir              if you wish to execute the command on every directory in repository.
  -o, --organisation string   organisation of the repository i.e. ministryofjustice (default "ministryofjustice")
  -r, --repository string     a blob of the repository name i.e. cloud-platform-terraform
  -s, --skip-commit           whether or not you want to create a commit and PR.

Global Flags:
      --config string   config file (default is $HOME/.cloud-platform-git-xargs.yaml)

```

## How to install it

These installation instructions are for a Mac. If you have a different kind of computer, please amend the steps appropriately.

Please substitute the latest release number. You can see the latest release number in the badge near the top of this page, and all available releases on this page.

```bash
RELEASE=<insert latest release>
wget https://github.com/ministryofjustice/cloud-platform-git-xargs/releases/download/${RELEASE}/cloud-platform-git-xargs_${RELEASE}_darwin_amd64.tar.gz
tar xzvf cloud-platform-git-xargs_${RELEASE}_darwin_amd64.tar.gz
mv cloud-platform-git-xargs /usr/local/bin/
```

NB: You may need to manually open the file to override OSX restrictions against executing binaries downloaded from the internet. To do this, locate the file in the Finder, right-click it and choose "Open". After doing this once, you should be able to run the command as normal.

## How to contribute

You will need Go installed. The makefile contains a build, test and release command. To release you must change the release variable to the latest tag and then run `make release`.

There are GitHub actions in this repository that will:

- build and release on new tag
- test code on PR and push
- format all codebase

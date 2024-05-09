# Contributing Guidelines
<BR>

## Umbrella Project
OpenEBS is an "umbrella project". Every project, repository and file in the OpenEBS organization adopts and follows the policies found in the Community repo umbrella project files.
<BR>

This project follows the [OpenEBS Contributor Guidelines](https://github.com/openebs/community/CONTRIBUTING.md)

# Contributing to zfs localpv

ZFS LocalPV uses the standard GitHub pull requests process to review and accept contributions.  There are several areas that could use your help. For starters, you could help in improving the sections in this document by either creating a new issue describing the improvement or submitting a pull request to this repository. The issues are maintained at [zfs-localpv/issues](https://github.com/openebs/zfs-localpv/issues) repository.

* If you are a first-time contributor, please see [Steps to Contribute](#steps-to-contribute).
* If you have documentation improvement ideas, go ahead and create a pull request. See [Pull Request checklist](#pull-request-checklist)
* If you would like to make code contributions, please start with [Setting up the Development Environment](#setting-up-your-development-environment).
* If you would like to work on something more involved, please connect with the OpenEBS Contributors. See [OpenEBS Community](https://github.com/openebs/openebs/tree/HEAD/community)

## Steps to Contribute

LocalPV-ZFS is an Apache 2.0 Licensed project and all your commits should be signed with Developer Certificate of Origin. See [Sign your work](#sign-your-work).

* Find an issue to work on or create a new issue. The issues are maintained at [zfs-localpv/issues](https://github.com/openebs/zfs-localpv/issues). You can pick up from a list of [good-first-issues](https://github.com/openebs/zfs-localpv/labels/good%20first%20issue).
* Claim your issue by commenting your intent to work on it to avoid duplication of efforts.
* Fork the repository on GitHub and clone.
* Create a branch from where you want to base your work (usually develop).
* Make your changes. If you are working on code contributions, please see [Setting up the Development Environment](#setting-up-your-development-environment).
* Relevant coding style guidelines are the [Go Code Review Comments](https://code.google.com/p/go-wiki/wiki/CodeReviewComments) and the _Formatting and style_ section of Peter Bourgon's [Go: Best Practices for Production Environments](http://peter.bourgon.org/go-in-production/#formatting-and-style).
* Commit your changes by making sure the commit messages convey the need and notes about the commit.
* Push your changes to the branch in your fork of the repository.
* Submit a pull request to the original repository. See [Pull Request checklist](#pull-request-checklist)


## Pull Request Checklist
* Rebase to the current develop branch before submitting your pull request.
* Commits should be as small as possible. Each commit should follow the checklist below:
  - For code changes, add tests relevant to the fixed bug or new feature. 
  - Before committing your code, make sure you have run `make format` and `make manifests`, to format the code and autogenerate the CRDs yaml.
  - Pass the compile and tests - includes spell checks, formatting, etc.
  - Commit header (first line) should convey what changed and it should follow the commit [guideline](https://github.com/openebs/openebs/blob/HEAD/contribute/git-commit-message.md)
  - Commit body should include details such as why the changes are required and how the proposed changes help
  - DCO Signed  
* If your PR is not getting reviewed or you need a specific person to review it, please reach out to the OpenEBS Contributors. See [OpenEBS Community](https://github.com/openebs/openebs/tree/HEAD/community)

---
## Sign your work

We use the Developer Certificate of Origin (DCO) as an additional safeguard for the OpenEBS projects. This is a well established and widely used mechanism to assure that contributors have confirmed their right to license their contribution under the project's license. Please read [dcofile](https://github.com/openebs/openebs/blob/HEAD/contribute/developer-certificate-of-origin). If you can certify it, then just add a line to every git commit message:

Please certify it by just adding a line to every git commit message. Any PR with Commits which does not have DCO Signoff will not be accepted:

````
  Signed-off-by: Random J Developer <random@developer.example.org>
````

Use your real name (sorry, no pseudonyms or anonymous contributions). The email id should match the email id provided in your GitHub profile.
If you set your `user.name` and `user.email` in git config, you can sign your commit automatically with `git commit -s`.

You can also use git [aliases](https://git-scm.com/book/tr/v2/Git-Basics-Git-Aliases) like `git config --global alias.ci 'commit -s'`. Now you can commit with `git ci` and the commit will be signed.

---
## Adding a changelog
If PR is about adding a new feature or bug fix then the Author of the PR is expected to add a changelog file with their pull request. This changelog file should be a new file created under the `changelogs/unreleased` folder. Name of this file must be in `pr_number-username` format and contents of the file should be the one-liner text which explains the feature or bug fix.

```sh
zfs-localpv/changelogs/unreleased   <- folder
    12-github_user_name            <- file
```
## Setting up your Development Environment

This project is implemented using Go and uses the standard golang tools for development and build. In addition, this project heavily relies on Docker and Kubernetes. It is expected that the contributors:
- are familiar with working with Go
- are familiar with Docker containers
- are familiar with Kubernetes and have access to a Kubernetes cluster or Minikube to test the changes.

For setting up a Development environment on your local host, see the detailed instructions [here](./docs/developer-setup.md).

The ZFS LocalPV design document is available [here](https://github.com/openebs/openebs/blob/HEAD/contribute/design/1.x/csi/20190805-csi-zfspv-volume-provisioning.md).

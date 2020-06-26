# Development Workflow

## Prerequisites

* You have Go 1.12.5 installed on your local host/development machine.
* You have Docker installed on your local host/development machine. Docker is required for building zfs-driver container images and to push them into a Kubernetes cluster for testing. 
* You have `kubectl` installed. For running integration tests, you can create a Minikube cluster on local host/development machine. Don't worry if you don't have access to the Kubernetes cluster, raising a PR with the zfs-localpv repository will run integration tests for your changes against a Minikube cluster.

## Initial Setup

### Fork in the cloud

1. Visit https://github.com/openebs/zfs-localpv
2. Click `Fork` button (top right) to establish a cloud-based fork.

### Clone fork to local host

Place openebs/zfs-localpv's code on your local machine using the following cloning procedure.
Create your clone:

# Note: Here user= your github profile name
git clone https://github.com/$user/zfs-localpv.git

# Configure remote upstream
cd zfs-localpv
git remote add upstream https://github.com/openebs/zfs-localpv.git

# Never push to upstream master
git remote set-url --push upstream no_push

# Confirm that your remotes make sense:
git remote -v
```

Install the build dependencies.
  * Run `make bootstrap` to install the required Go tools

## Git Development Workflow

### Always sync your local repository:
Open a terminal on your local host. Change directory to the zfs-localpv fork root.

```sh
$ cd zfs-localpv
```

 Checkout the master branch.

 ```sh
 $ git checkout master
 Switched to branch 'master'
 Your branch is up-to-date with 'origin/master'.
 ```

 Recall that origin/master is a branch on your remote GitHub repository.
 Make sure you have the upstream remote openebs/zfs-localpv by listing them.

 ```sh
 $ git remote -v
 origin	https://github.com/$user/zfs-localpv.git (fetch)
 origin	https://github.com/$user/zfs-localpv.git (push)
 upstream	https://github.com/openebs/zfs-localpv.git (fetch)
 upstream	https://github.com/openebs/zfs-localpv.git (no_push)
 ```

 If the upstream is missing, add it by using below command.

 ```sh
 $ git remote add upstream https://github.com/openebs/zfs-localpv.git
 ```
 Fetch all the changes from the upstream master branch.

 ```sh
 $ git fetch upstream master
 remote: Counting objects: 141, done.
 remote: Compressing objects: 100% (29/29), done.
 remote: Total 141 (delta 52), reused 46 (delta 46), pack-reused 66
 Receiving objects: 100% (141/141), 112.43 KiB | 0 bytes/s, done.
 Resolving deltas: 100% (79/79), done.
 From github.com:openebs/zfs-localpv
   * branch            master     -> FETCH_HEAD
 ```

 Rebase your local master with the upstream/master.

 ```sh
 $ git rebase upstream/master
 First, rewinding head to replay your work on top of it...
 Fast-forwarded master to upstream/master.
 ```
 This command applies all the commits from the upstream master to your local master.

 Check the status of your local branch.

 ```sh
 $ git status
 On branch master
 Your branch is ahead of 'origin/master' by 38 commits.
 (use "git push" to publish your local commits)
 nothing to commit, working directory clean
 ```
 Your local repository now has all the changes from the upstream remote. You need to push the changes to your own remote fork which is origin master.

 Push the rebased master to origin master.

 ```sh
 $ git push origin master
 Username for 'https://github.com': $user
 Password for 'https://$user@github.com':
 Counting objects: 223, done.
 Compressing objects: 100% (38/38), done.
 Writing objects: 100% (69/69), 8.76 KiB | 0 bytes/s, done.
 Total 69 (delta 53), reused 47 (delta 31)
 To https://github.com/$user/zfs-localpv.git
 8e107a9..5035fa1  master -> master
 ```

### Contributing to a feature or bugfix. 

Always start with creating a new branch from master to work on a new feature or bugfix. Your branch name should have the format XX-descriptive where XX is the issue number you are working on followed by some descriptive text. For example:

 ```sh
 $ git checkout master
 # Make sure the master is rebased with the latest changes as described in previous step.
 $ git checkout -b 1234-fix-developer-docs
 Switched to a new branch '1234-fix-developer-docs'
 ```
Happy Hacking!

### Building and Testing your changes

* run `make` in the top directory. It will:
  * Build the binary.
  * Build the docker image with the binary.

* Test your changes
  * setup the ZFS POOL on the nodes, check Setup in [readme](../README.md).
  * Integration tests are written in ginkgo and run against a minikube cluster. Minikube cluster should be running so as to execute the tests. To install minikube follow the doc [here](https://kubernetes.io/docs/tasks/tools/install-minikube/). 
  * `sudo -E env "PATH=$PATH" make ci` execute the integration tests

### Keep your branch in sync

[Rebasing](https://git-scm.com/docs/git-rebase) is very import to keep your branch in sync with the changes being made by others and to avoid huge merge conflicts while raising your Pull Requests. You will always have to rebase before raising the PR. 

```sh
# While on your myfeature branch (see above)
git fetch upstream
git rebase upstream/master
```

While you rebase your changes, you must resolve any conflicts that might arise and build and test your changes using the above steps. 

## Submission

### Create a pull request

Before you raise the Pull Requests, ensure you have reviewed the checklist in the [CONTRIBUTING GUIDE](../CONTRIBUTING.md):
- Ensure that you have re-based your changes with the upstream using the steps above.
- Ensure that you have added the required unit tests for the bug fixes or new feature that you have introduced. 
- Ensure your commits history is clean with proper header and descriptions.

Go to the [openebs/zfs-localpv github](https://github.com/openebs/zfs-localpv) and follow the Open Pull Request link to raise your PR from your development branch.




# Development Workflow

## Prerequisites

### System Packages

* You need to have a docker service installed on your local host/development machine. Docker is required for building zfs-driver container images and to push them into a Kubernetes cluster for testing.
* The host/development machine must also run a local single node kubernetes cluster in order to run the ci tests. \
  But don't worry if this is not possible, raising a PR with the zfs-localpv repository will run integration tests for your changes against a Minikube cluster.

### [Nix][nix]

This is the recommend way of setting up a development environment.
Usually [Nix][nix-install] can be installed via (Do **not** use `sudo`!):

```bash
curl -L https://nixos.org/nix/install | sh
```

> **Can't install Nix, or don't want to?**
>
> That's totally fine, you'll need the following additional system packages:
>
> * `Go` 1.24
> * `kubectl`

### Fork in the cloud

1. Visit https://github.com/openebs/zfs-localpv
2. Click `Fork` button (top right) to establish a cloud-based fork.

### Clone fork to local host

Place openebs/zfs-localpv's code on your local machine using the following cloning procedure:

```sh
$ mkdir -p ~/git
$ cd ~/git
# $user is your github user
$ git clone https://github.com/$user/zfs-localpv.git # you may use ssh instead
$ cd zfs-localpv
$ git remote add upstream https://github.com/openebs/zfs-localpv.git
# Don't push to upstream directly
$ git remote set-url --push upstream no_push
```

## Git Development Workflow

### Always sync your local repository

Checkout the develop branch.

```sh
$ cd ~/git/zfs-localpv
$ git checkout develop
Switched to branch 'develop'
Your branch is up-to-date with 'origin/develop'.
```

Recall that origin/develop is a branch on your remote GitHub repository.
Make sure you have the upstream remote openebs/zfs-localpv by listing them.

```sh
$ git remote -v
origin https://github.com/$user/zfs-localpv.git (fetch)
origin https://github.com/$user/zfs-localpv.git (push)
upstream https://github.com/openebs/zfs-localpv.git (fetch)
upstream https://github.com/openebs/zfs-localpv.git (no_push)
```

If the upstream is missing, add it by using below command.

```sh
git remote add upstream https://github.com/openebs/zfs-localpv.git
```

Fetch all the changes from the upstream develop branch.

```sh
$ git fetch upstream develop
remote: Counting objects: 141, done.
remote: Compressing objects: 100% (29/29), done.
remote: Total 141 (delta 52), reused 46 (delta 46), pack-reused 66
Receiving objects: 100% (141/141), 112.43 KiB | 0 bytes/s, done.
Resolving deltas: 100% (79/79), done.
From github.com:openebs/zfs-localpv
  * branch            develop     -> FETCH_HEAD
```

Rebase your local develop with the upstream/develop.

```sh
$ git rebase upstream/develop
First, rewinding head to replay your work on top of it...
Fast-forwarded develop to upstream/develop.
```

This command applies all the commits from the upstream develop to your local develop.

Check the status of your local branch.

```sh
$ git status
On branch develop
Your branch is ahead of 'origin/develop' by 38 commits.
(use "git push" to publish your local commits)
nothing to commit, working directory clean
```

Your local repository now has all the changes from the upstream remote. You need to push the changes to your own remote fork which is origin develop.

Push the rebased develop to origin develop.

```sh
$ git push origin develop
Username for 'https://github.com': $user
Password for 'https://$user@github.com':
Counting objects: 223, done.
Compressing objects: 100% (38/38), done.
Writing objects: 100% (69/69), 8.76 KiB | 0 bytes/s, done.
Total 69 (delta 53), reused 47 (delta 31)
To https://github.com/$user/zfs-localpv.git
8e107a9..5035fa1  develop -> develop
```

### Contributing to a feature or bugfix

Always start with creating a new branch from develop to work on a new feature or bugfix. Your branch name should have the format XX-descriptive where XX is the issue number you are working on followed by some descriptive text. For example:

 ```sh
 $ git checkout develop
 # Make sure the develop is rebased with the latest changes as described in previous step.
 $ git checkout -b 1234-fix-developer-docs
 Switched to a new branch '1234-fix-developer-docs'
 ```

Happy Hacking!

### Keep your branch in sync

[Rebasing](https://git-scm.com/docs/git-rebase) is very import to keep your branch in sync with the changes being made by others and to avoid huge merge conflicts while raising your Pull Requests. You will always have to rebase before raising the PR.

```sh
# While on your myfeature branch (see above)
$ git fetch upstream
$ git rebase upstream/develop
```

While you rebase your changes, you must resolve any conflicts that might arise and build and test your changes using the above steps.

## Building

Before starting, ensure you have installed all the dependencies.
If you're using [nix], then simply enter the [nix-shell] which will setup a shell with all required packages:

NOTE: If you're using zsh plugin in nix-shell, make sure your zshrc configurations don't override the Go envs set by the nix-shell.

```sh
$ nix-shell
go: downloading ....
....
```

> *NOTE*: you can keep using your favorite shell by running it, example: `nix-shell --run zsh`.

On non-nix environments, run `make bootstrap` to install the required Go tools.

We have several `make` commands available at your convenience:

* `make bootstrap`: as seen previously, installs all the required Go tools.
* `make format`: will ensure the code is formatted according to the the guidelines.
* `make golint`: comprehensive linter that helps catch potential issues such as coding style problems, possible bugs, and performance issues.
* `make clean`: cleans all the installed Go tools, intermediate and generated artifacts.
* `make`: builds the binary and a test docker image.

There are more commands, take your time to read through the the [Makefile](../Makefile) and [Makefile-buildx](../Makefile.buildx.mk).

## Testing

Simple unit testing can be done via make:

```sh
$ make test
--> Running go fmt
--> Running go test
?       github.com/openebs/zfs-localpv/cmd      [no test files]
ok      github.com/openebs/zfs-localpv/pkg/driver       0.011s  coverage: 0.5% of statements
?       github.com/openebs/zfs-localpv/pkg/equality     [no test files]
...
```

The ci tests perform a more comprehensive testing, we need to make use of a [Kubernetes][K8s] cluster. \
Sadly at this point in time, the ci tests expect to be running on a throw-away single node [K8s] cluster. \
This means your development/testing system must be running on this very same node. \

### Kubernetes Clusters

We suggest a few options for this:

* [nixos-shell]
* [K3d]
* [minikube]
* [kind]

#### [NixOs-Shell][nixos-shell]

If you're already using [nix-shell], then why not take this option? \
How does it work? \
It spawns a headless qemu virtual machines based on a [configuration file](../vm.nix) and it provides console access in the same terminal window.

The provided [configuration file](../vm.nix) deploys as a single node [K3s] cluster with the required zfs driver and tools.

If you're already on the `nix-shell`, simply run it:

```sh
$ nixos-shell
Welcome to NixOS 24.11 (Vicuna)!

[  OK  ] Created slice Slice /system/getty.
[  OK  ] Created slice Slice /system/modprobe.
[  OK  ] Created slice Slice /system/serial-getty
....
[  OK  ] Reached target Network is Online.
         Starting Docker Application Container Engine...
         Starting k3s service...
[  OK  ] Started Docker Application Container Engine.

<<< Welcome to NixOS 24.11.712512.3f0a8ac25fb6 (x86_64) - hvc0 >>>
Log in as "root" with an empty password.
Run 'nixos-help' for the NixOS manual.
nixos login: root

```

Simply log in with the root user (no password) get hacking!
To leave this shell you can use the key combination `Ctrl-a x`.
You can then enter the [nixos-shell] again using the same command.
If you've made irreparable damage to the virtual machine, simply delete it and start anew. This is as simple as:

```sh
# Leave the vm terminal and run:
rm nixos.qcow2
nixos-shell
```

#### MiniKube

To install minikube follow the doc [here](https://kubernetes.io/docs/tasks/tools/install-minikube/).

> *NOTE*: the minikube cluster must be started with the [none driver](https://minikube.sigs.k8s.io/docs/drivers/none/)

### Running the tests

Integration tests are written in ginkgo and run against a [K8s] cluster.

> *NOTE*: For nixos-shell, remember to run these command within the nixos-shell vm!

Then you can run the tests:

```sh
./ci/ci-test.sh run
```

> *WARNING*: Each individual tests don't currently clean up after themselves properly, so a failed test may affect subsequent tests and even new test runs!

If the test script was killed before cleaning up, you can issue a cleanup as such:

```sh
./ci/ci-test.sh clean
```

You can also request one before running the tests:

```sh
./ci/ci-test.sh run --reset
```

> *WARNING*: If you modify the code, remember to rebuild and load the new image, example:
>
> ```sh
> ./ci/ci-test.sh run --build-always
> ```

If this doesn't work, you might need to dig a little deeper, or, if you're on nixos-shell, simply start over again:

```sh
rm nixos.qcow2
nixos-shell
```

And there you have it, a clean VM ready to be broken again :)

## Submission

### Create a pull request

Before you raise the Pull Requests, ensure you have reviewed the checklist in the [CONTRIBUTING GUIDE](../CONTRIBUTING.md):

* Ensure that you have re-based your changes with the upstream using the steps above.
* Ensure that you have added the required unit tests for the bug fixes or new feature that you have introduced.
* Ensure your commits history is clean with proper header and descriptions.

Go to the [openebs/zfs-localpv github](https://github.com/openebs/zfs-localpv) and follow the Open Pull Request link to raise your PR from your development branch.

[nix]: https://nixos.org/
[nix-install]: https://nixos.org/download.html
[nix-shell]: https://nixos.org/guides/nix-pills/10-developing-with-nix-shell.html
[K8s]: https://kubernetes.io/
[minikube]: https://minikube.sigs.k8s.io/docs/
[nixos-shell]: https://github.com/Mic92/nixos-shell
[kind]: https://kind.sigs.k8s.io/
[K3d]: https://k3d.io/stable/
[K3s]: https://k3s.io/

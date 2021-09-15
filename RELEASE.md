# Release Process
zfs-localpv follows a monthly release cadence. The scope of the release is determined by contributor availability. The scope is published in the [Release Tracker Projects](https://github.com/orgs/openebs/projects/10).

## Release Candidate Verification Checklist

Every release has release candidate builds that are created starting from the third week into the release. These release candidate builds help to freeze the scope and maintain the quality of the release. The release candidate builds will go through:
- Platform Verification
- Regression and Feature Verification Automated tests.
- Exploratory testing by QA engineers
- Strict security scanners on the container images
- Upgrade from previous releases
- Beta testing by users on issues that they are interested in.
- Dogfooding on OpenEBS workload and e2e infrastructure clusters.

If any issues are found during the above stages, they are fixed and a new release candidate builds are generated.

Once all the above tests are completed, a main release tagged image is published.

## Release Tagging

zfs-localpv is released as container image with versioned tag.

Before creating a release, the repo owner needs to create a separate branch from the active branch, which is `develop`. Name of the branch should follow the naming convention of `v.0.7.x` if release is for 0.7.0.

Once the release branch is created, changelog from `changelogs/unreleased` needs to be moved to release specific folder, if release branch is `v0.7.x` then folder will be `changelogs/v0.7.x`.

The format of the release tag is either "Release-Name-RC1" or "Release-Name" depending on whether the tag is a release candidate or a release. (Example: 0.7.0-RC1 is a GitHub release tag for zfs-localpv release candidate build. 0.7.0 is the release tag that is created after the release criteria are satisfied by the zfs-localpv builds.)

Once the release is triggered, Travis build process has to be monitored. Once Travis build passes, images are pushed to docker hub and quay.io. Images can be verified by going through docker hub and quay.io. Also, the images shouldn't have any high level vulnerabilities.

Images are published at following location :

https://quay.io/repository/openebs/zfs-driver?tab=tags
https://hub.docker.com/r/openebs/zfs-driver/tags

Once a release is created, update the release description with the changelog mentioned in `changelog/v0.7.x`. Once the changelogs are updated in the release, the repo owner needs to create a PR to `develop` with the following details:
1. update the changelog from `changelog/v0.7.x` to `zfs-localpv/CHANGELOG.md`
2. If a release is not an RC tag then PR should include the changes to remove `changelog/v0.7.x` folder.
3. If a release is an RC tag then PR should include the changes to remove the changelog from `changelog/v0.7.x` which are already mentioned in `zfs-localpv/CHANGELOG.md` as part of step number 1.

# Build zfs-driver docker images with buildx
# Experimental docker feature to build cross platform multi-architecture docker images
# https://docs.docker.com/buildx/working-with-buildx/

ifeq (${TAG}, )
  export TAG=ci
endif

# default list of platforms for which multiarch image is built
ifeq (${PLATFORMS}, )
	export PLATFORMS="linux/amd64,linux/arm64"
endif

# if IMG_RESULT is unspecified, by default the image will be pushed to registry
ifeq (${IMG_RESULT}, load)
	export PUSH_ARG="--load"
    # if load is specified, image will be built only for the build machine architecture.
    export PLATFORMS="local"
else ifeq (${IMG_RESULT}, cache)
	# if cache is specified, image will only be available in the build cache, it won't be pushed or loaded
	# therefore no PUSH_ARG will be specified
else
	export PUSH_ARG="--push"
endif

# Name of the multiarch image for csi-driver
DOCKERX_IMAGE_CSI_DRIVER:=${IMAGE_ORG}/zfs-driver:${TAG}

.PHONY: docker.buildx
docker.buildx:
	export DOCKER_CLI_EXPERIMENTAL=enabled
	@if ! docker buildx ls | grep -q container-builder; then\
		docker buildx create --platform ${PLATFORMS} --name container-builder --use;\
	fi
	@docker buildx build --platform "${PLATFORMS}" \
		-t "$(DOCKERX_IMAGE_NAME)" ${BUILD_ARGS} \
		-f $(PWD)/buildscripts/$(COMPONENT)/$(COMPONENT).Dockerfile \
		. ${PUSH_ARG}
	@echo "--> Build docker image: $(DOCKERX_IMAGE_NAME)"
	@echo

.PHONY: buildx.csi-driver
buildx.csi-driver:
	@echo '--> Building csi-driver binary...'
	@pwd
	@PNAME=${CSI_DRIVER} CTLNAME=${CSI_DRIVER} BUILDX=true sh -c "'$(PWD)/buildscripts/build.sh'"
	@echo '--> Built binary.'
	@echo

.PHONY: docker.buildx.csi-driver
docker.buildx.csi-driver: DOCKERX_IMAGE_NAME=$(DOCKERX_IMAGE_CSI_DRIVER)
docker.buildx.csi-driver: COMPONENT=$(CSI_DRIVER)
docker.buildx.csi-driver: BUILD_ARGS=$(DBUILD_ARGS)
docker.buildx.csi-driver: docker.buildx


.PHONY: buildx.push.csi-driver
buildx.push.csi-driver:
	BUILDX=true DIMAGE=${IMAGE_ORG}/zfs-driver ./buildscripts/push

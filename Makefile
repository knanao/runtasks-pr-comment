BUILD_VERSION=v0.1.1
REPOSITORY=ghcr.io/knanao/runtasks-pr-comment

.PHONY: build
build:
	docker build --platform linux/amd64 -t ${REPOSITORY}:${BUILD_VERSION} -t ${REPOSITORY}:latest .

.PHONY: push
push: build
push:
	docker push ${REPOSITORY}:${BUILD_VERSION}
	docker push ${REPOSITORY}:latest

.PHONY: run
run: GITHUB_OAUTH_TOKEN ?=
run: GITHUB_APP_ID ?=
run: GITHUB_APP_PRIVATE_KEY ?=
run: GITHUB_APP_INSTALLATION_ID ?=
run: TFC_RUN_TASK_HMAC_KEY ?=
run:
ifndef ${GITHUB_OAUTH_TOKEN}
	docker run -e TFC_RUN_TASK_HMAC_KEY=${TFC_RUN_TASK_HMAC_KEY} -e GITHUB_OAUTH_TOKEN=${GITHUB_OAUTH_TOKEN} ${REPOSITORY}:${BUILD_VERSION}
else
	docker run -e TFC_RUN_TASK_HMAC_KEY=${TFC_RUN_TASK_HMAC_KEY} -e GITHUB_APP_ID=${GITHUB_APP_ID} -e GITHUB_APP_PRIVATE_KEY=${GITHUB_APP_PRIVATE_KEY} -e GITHUB_APP_INSTALLATION_ID=${GITHUB_APP_INSTALLATION_ID} ${REPOSITORY}:${BUILD_VERSION}
endif

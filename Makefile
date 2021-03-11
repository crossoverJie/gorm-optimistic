SHELL := /bin/bash

RELEASE_VERSION = v0.0.2

git-tag:
	git tag $(RELEASE_VERSION); \
	git push origin $(RELEASE_VERSION); \

download:
	$Q go mod download
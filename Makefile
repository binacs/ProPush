BUILD_TAGS?=
BUILD_FLAGS = -ldflags "-X github.com/BinacsLee/ProPush/version.GitCommit=`git rev-parse HEAD`"

all: build install

build:
	go build $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' -o bin/propush ./cmd

install:
	go install $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' ./cmd



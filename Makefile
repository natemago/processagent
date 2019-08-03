LD_FLAGS=-ldflags "-s -w"

build_linux_amd64: export GOOS=linux
build_linux_amd64: export GOARCH=amd64
build_linux_amd64: 
	go build $(LD_FLAGS) -o "build/processagent_${GOOS}_${GOARCH}" ./main

build_darwin_amd64: export GOOS=darwin
build_darwin_amd64: export GOARCH=amd64
build_darwin_amd64:
	go build $(LD_FLAGS) -o "build/processagent_${GOOS}_${GOARCH}" ./main

build_linux_arm: export GOOS=linux
build_linux_arm: export GOARCH=arm
build_linux_arm:
	go build $(LD_FLAGS) -o "build/processagent_${GOOS}_${GOARCH}" ./main

build_linux_arm64: export GOOS=linux
build_linux_arm64: export GOARCH=arm64
build_linux_arm64:
	go build $(LD_FLAGS) -o "build/processagent_${GOOS}_${GOARCH}" ./main

all: build_linux_amd64 build_darwin_amd64 build_linux_arm build_linux_arm64

clean:
	rm -rf build
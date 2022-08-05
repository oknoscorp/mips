BINARY_NAME=bin/mips.out

build:
	go build -o ${BINARY_NAME}

run:
	./${BINARY_NAME}

all: build run
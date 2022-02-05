ROOTPROJECT ?= ../root
APIPROJECT = .
include ${ROOTPROJECT}/include.mk

# Dummy targets for cluster/up and cluster/teardown
.PHONY: up down

up:
down:
build: test
# Test code
.PHONY: test

test: go/test

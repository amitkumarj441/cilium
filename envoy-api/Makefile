# Copyright 2017 Authors of Cilium
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

include ../Makefile.defs

PROTOC ?= protoc
PROTO_PATH = envoy-api
GO_OUT = ../pkg/envoy

API_PROTOS := api/*.proto
API_GOS = $(API_PROTOS:.proto=.pb.go)
API_SOURCES := $(addprefix $(PROTO_PATH)/,$(API_PROTOS))
API_TARGETS := $(addprefix $(GO_OUT)/,$(API_GOS))

FILTER_PROTOS := api/filter/http_connection_manager.proto
FILTER_GOS = $(FILTER_PROTOS:.proto=.pb.go)
FILTER_SOURCES := $(addprefix $(PROTO_PATH)/,$(FILTER_PROTOS))
FILTER_TARGETS := $(addprefix $(GO_OUT)/,$(FILTER_GOS))

GO_TARGET = test
GO_SOURCES := $(shell find . -name '*.go')

all: submodule_update $(API_TARGETS) $(FILTER_TARGETS) $(GO_TARGET)

submodule_update: force
	git submodule update

$(GO_TARGET): $(GO_SOURCES)
	$(GO) build -i $(GOBUILD) -o $(GO_TARGET)

$(GO_OUT):
	mkdir $(GO_OUT)

$(API_TARGETS): $(API_SOURCES) | $(GO_OUT)
	protoc --proto_path=$(PROTO_PATH) -I google --go_out=plugins=grpc:$(GO_OUT) $(API_SOURCES)

$(API_SOURCES): $(PROTO_PATH)

$(FILTER_TARGETS): $(FILTER_SOURCES) | $(GO_OUT)
	protoc --proto_path=$(PROTO_PATH) -I google --go_out=plugins=grpc:$(GO_OUT) $(FILTER_SOURCES)

clean: force
	-rm -f $(GO_TARGET)
	-$(GO) clean
	-rm $(API_TARGETS) $(FILTER_TARGETS)
	-rmdir $(GO_OUT)/api/filter
	-rmdir $(GO_OUT)/api

.PHONY: force
force :;

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

TARGET= ./bazel-bin/envoy
DEBUG_TARGET=./bazel-out/local-dbg/bin/envoy

BAZEL ?= bazel
CLANG ?= clang
CLANG_FORMAT ?= clang-format
BUILDIFIER ?= buildifier
STRIP ?= strip

all: submodule_update $(TARGET) $(DEBUG_TARGET)

submodule_update: force
	git submodule update

$(TARGET): force
	$(BAZEL) build -c opt //:envoy

$(DEBUG_TARGET): force
	$(BAZEL) build -c dbg //:envoy

install: force
	$(INSTALL) -m 0755 -d $(DESTDIR)$(BINDIR)
	-rm $(DESTDIR)$(BINDIR)/cilium-envoy
	$(INSTALL) -m 0755 -T $(TARGET) $(DESTDIR)$(BINDIR)/cilium-envoy
	$(STRIP) $(DESTDIR)$(BINDIR)/cilium-envoy
	-rm $(DESTDIR)$(BINDIR)/cilium-envoy-debug
	$(INSTALL) -m 0755 -T $(DEBUG_TARGET) $(DESTDIR)$(BINDIR)/cilium-envoy-debug

clean: force
	$(BAZEL) clean

check: force
	CLANG_FORMAT=$(CLANG_FORMAT) BUILDIFIER=$(BUILDIFIER) envoy/tools/check_format.py --add-excluded-prefixes="./envoy/" check

fix: force
	CLANG_FORMAT=$(CLANG_FORMAT) BUILDIFIER=$(BUILDIFIER) envoy/tools/check_format.py --add-excluded-prefixes="./envoy/" fix

tests: all force
	./envoy_binary_test.sh
	$(BAZEL) test //:cilium_integration_test

.PHONY: force
force :;

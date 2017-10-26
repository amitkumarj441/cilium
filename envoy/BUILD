package(default_visibility = ["//visibility:public"])

load(
    "@envoy//bazel:envoy_build_system.bzl",
    "envoy_cc_binary",
    "envoy_cc_library",
    "envoy_cc_test",
)

envoy_cc_binary(
    name = "envoy",
    repository = "@envoy",
    deps = [
        ":bpf_metadata_lib",
        ":cilium_l7_lib",
        "@envoy//source/exe:envoy_main_entry_lib",
    ],
)

envoy_cc_library(
    name = "cilium_l7_lib",
    srcs = ["cilium_l7.cc"],
    hdrs = ["cilium_l7.h"],
    repository = "@envoy",
    deps = [
        "@envoy//source/exe:envoy_common_lib",
    ],
)

envoy_cc_library(
    name = "bpf_metadata_lib",
    srcs = ["bpf_metadata.cc", "bpf.cc", "proxymap.cc"],
    hdrs = ["bpf_metadata.h", "bpf.h", "proxymap.h"],
    repository = "@envoy",
    deps = [
        "@envoy//include/envoy/buffer:buffer_interface",
        "@envoy//include/envoy/network:connection_interface",
        "@envoy//include/envoy/network:filter_interface",
        "@envoy//include/envoy/registry:registry",
        "@envoy//include/envoy/server:filter_config_interface",
        "@envoy//source/common/common:assert_lib",
        "@envoy//source/common/common:logger_lib",
        "@envoy//source/common/network:address_lib",
    ],
)

envoy_cc_test(
    name = "cilium_integration_test",
    srcs = ["cilium_integration_test.cc"],
    data =  ["cilium_proxy_test.json"],
    repository = "@envoy",
    deps = [
        ":bpf_metadata_lib",
        ":cilium_l7_lib",
        "@envoy//test/integration:integration_lib"
    ],
)

sh_test(
    name = "envoy_binary_test",
    srcs = ["envoy_binary_test.sh"],
    data = [":envoy"],
)

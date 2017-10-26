#pragma once

#include <string>

#include "envoy/stats/stats_macros.h"
#include "server/config/network/http_connection_manager.h"

#include "common/common/logger.h"

namespace Envoy {
namespace Cilium {

/**
 * All router filter stats. @see stats_macros.h
 */
// clang-format off
#define ALL_CILIUM_STATS(COUNTER)                                                                  \
  COUNTER(access_denied)                                                                           \
// clang-format on

/**
 * Struct definition for all router filter stats. @see stats_macros.h
 */
struct FilterStats {
  ALL_CILIUM_STATS(GENERATE_COUNTER_STRUCT)
};

/**
 * Global configuration for Cilium HTTP filter.  This
 * represents all global state shared among the working thread
 * instances of the filter.
 */
class Config {
public:
  Config(const Json::Object& config, Stats::Scope& scope);

  FilterStats stats_;
};

typedef std::shared_ptr<Config> ConfigSharedPtr;

// Each request gets their own instance of this DecoderFilter, and
// they can run parallel from multiple worker threads, all accessing
// the shared configuration.
class DecoderFilter : Logger::Loggable<Logger::Id::router>,
  public Http::StreamDecoderFilter {
public:
  DecoderFilter(ConfigSharedPtr& config) : config_(config) {}

  // Http::StreamFilterBase
  void onDestroy() override;

  // Http::StreamDecoderFilter
  Http::FilterHeadersStatus decodeHeaders(Http::HeaderMap& headers, bool) override;
  Http::FilterDataStatus decodeData(Buffer::Instance&, bool) override;
  Http::FilterTrailersStatus decodeTrailers(Http::HeaderMap&) override;
  void setDecoderFilterCallbacks(Http::StreamDecoderFilterCallbacks& callbacks) override;

private:
  ConfigSharedPtr config_;
  Http::StreamDecoderFilterCallbacks* callbacks_;
  const Http::LowerCaseString& headerKey();
  const std::string& headerValue();
};

} // Cilium
} // Envoy

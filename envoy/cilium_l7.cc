#include <string>

#include "cilium_l7.h"

#include "envoy/registry/registry.h"

#include "common/common/enum_to_int.h"

#include "server/config/network/http_connection_manager.h"

namespace Envoy {
namespace Cilium {

class ConfigFactory : public Server::Configuration::NamedHttpFilterConfigFactory {
public:
  Server::Configuration::HttpFilterFactoryCb createFilterFactory(const Json::Object& json, const std::string&,
                                          Server::Configuration::FactoryContext& context) override {
    Cilium::ConfigSharedPtr config(new Cilium::Config(json, context.scope()));

    return [config](Http::FilterChainFactoryCallbacks& callbacks) mutable -> void {
      callbacks.addStreamDecoderFilter(std::make_shared<Cilium::DecoderFilter>(config));
    };
  }
  std::string name() override { return "cilium_l7"; }
  Server::Configuration::HttpFilterType type() override { return Server::Configuration::HttpFilterType::Decoder; }
};

/**
 * Static registration for this sample filter. @see RegisterFactory.
 */
static Registry::RegisterFactory<ConfigFactory, Server::Configuration::NamedHttpFilterConfigFactory>
    register_;

Config::Config(const Json::Object& config, Stats::Scope& scope)
  : stats_{ALL_CILIUM_STATS(POOL_COUNTER_PREFIX(scope, "cilium"))} {
  UNREFERENCED_PARAMETER(config);
}

void DecoderFilter::onDestroy() {}

Http::FilterHeadersStatus DecoderFilter::decodeHeaders(Http::HeaderMap& headers, bool) {
  // Cilium configures security policy on route entries, whitelisting
  // allowed traffic. Return 403 if no route is found.
  auto route = callbacks_->route();
  if (!route) {
    config_->stats_.access_denied_.inc();
    ENVOY_STREAM_LOG(debug, "access denied for URL '{}'", *callbacks_,
                     headers.Path()->value().c_str());

    // Add new ResponseFlag for "adminitratively denied"?
    Http::HeaderMapPtr response_headers{new Http::HeaderMapImpl{
        {Http::Headers::get().Status, std::to_string(enumToInt(Http::Code::Forbidden))}}};
    callbacks_->encodeHeaders(std::move(response_headers), true);
    return Http::FilterHeadersStatus::StopIteration;
  }

  return Http::FilterHeadersStatus::Continue;
}

Http::FilterDataStatus DecoderFilter::decodeData(Buffer::Instance&, bool) {
  return Http::FilterDataStatus::Continue;
}

Http::FilterTrailersStatus DecoderFilter::decodeTrailers(Http::HeaderMap&) {
  return Http::FilterTrailersStatus::Continue;
}

void DecoderFilter::setDecoderFilterCallbacks(Http::StreamDecoderFilterCallbacks& callbacks) {
  callbacks_ = &callbacks;
}

} // Cilium
} // Envoy

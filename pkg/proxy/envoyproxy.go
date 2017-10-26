// Copyright 2016-2017 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package proxy

import (
	"fmt"
	"sync"

	"github.com/cilium/cilium/pkg/envoy"
	"github.com/cilium/cilium/pkg/policy"

	"github.com/spf13/viper"
)

// the global Envoy instance
var envoyProxy *envoy.Envoy

// EnvoyRedirect implements the Redirect interface for a l7 proxy
type EnvoyRedirect struct {
	id     string
	toPort uint16
}

// ToPort returns the redirect port of an OxyRedirect
func (r *EnvoyRedirect) ToPort() uint16 {
	return r.toPort
}

var envoyOnce sync.Once

// createOxyRedirect creates a redirect with corresponding proxy
// configuration. This will launch a proxy instance.
func createEnvoyRedirect(l4 *policy.L4Filter, id string, source ProxySource, to uint16) (Redirect, error) {
	envoyOnce.Do(func() {
		// Start Envoy on first invocation
		envoyProxy = envoy.StartEnvoy(true, 0, viper.GetString("state-dir"), viper.GetString("state-dir"))
	})

	if envoyProxy != nil {
		redir := &EnvoyRedirect{
			id:     id,
			toPort: to,
		}

		envoyProxy.AddListener(id, to, l4.L7RulesPerEp, l4.Ingress)

		return redir, nil
	}

	return nil, fmt.Errorf("%s: Envoy proxy process failed to start, can not add redirect ", id)
}

// UpdateRules replaces old l7 rules of a redirect with new ones.
func (r *EnvoyRedirect) UpdateRules(l4 *policy.L4Filter) error {
	if envoyProxy != nil {
		envoyProxy.UpdateListener(r.id, l4.L7RulesPerEp)
		return nil
	}
	return fmt.Errorf("%s: Envoy proxy process failed to start, can not update redirect ", r.id)
}

// Close the redirect.
func (r *EnvoyRedirect) Close() {
	if envoyProxy != nil {
		envoyProxy.RemoveListener(r.id)
	}
}

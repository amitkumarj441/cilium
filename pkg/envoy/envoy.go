package envoy

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/cilium/cilium/pkg/policy"

	log "github.com/sirupsen/logrus"
)

// Envoy manages a running Envoy proxy instance via the
// ListenerDiscoveryService and RouteDiscoveryService gRPC APIs.
type Envoy struct {
	cmd     *exec.Cmd
	LogPath string
	ldsSock string
	lds     *LDSServer
	rdsSock string
	rds     *RDSServer
}

func createConfig(filePath string, adminAddress string) {
	config := string("{\n" +
		"  \"listeners\": [],\n" +
		"  \"admin\": { \"access_log_path\": \"/dev/null\",\n" +
		"             \"address\": \"tcp://" + adminAddress + "\" },\n" +
		"  \"cluster_manager\": {\n" +
		"    \"clusters\": []\n" +
		"  }\n" +
		"}\n")

	log.Debug("Envoy: Configuration file: ", config)
	err := ioutil.WriteFile(filePath, []byte(config), 0644)
	if err != nil {
		log.WithError(err).Fatal("Envoy: Failed writing configuration file ", filePath)
	}
}

// StartEnvoy starts an Envoy proxy instance. If 'debug' is true, an
// debug version of the Envoy binary is started with the log level
// 'debug', otherwise a production version is started at the default
// log level.
func StartEnvoy(debug bool, adminPort int, stateDir, logDir string) *Envoy {
	bootstrapPath := filepath.Join(stateDir, "bootstrap.pb")
	configPath := filepath.Join(stateDir, "envoy-config.json")
	logPath := filepath.Join(logDir, "cilium-envoy.log")
	adminAddress := "127.0.0.1:" + strconv.Itoa(adminPort)
	ldsPath := filepath.Join(stateDir, "lds.sock")
	rdsPath := filepath.Join(stateDir, "rds.sock")

	e := &Envoy{LogPath: logPath, ldsSock: ldsPath, rdsSock: rdsPath}

	// Create configuration
	createBootstrap(bootstrapPath, "envoy1", "cluster1", "version1",
		"ldsCluster", ldsPath, "rdsCluster", rdsPath, "cluster1")
	createConfig(configPath, adminAddress)

	if debug {
		e.cmd = exec.Command("sh", "-c", "cilium-envoy-debug >"+logPath+" 2>&1 -l debug -c "+configPath+" -b "+bootstrapPath)
	} else {
		e.cmd = exec.Command("sh", "-c", "cilium-envoy >"+logPath+" 2>&1 -c "+configPath+" -b "+bootstrapPath)
	}

	log.Debug("Envoy: Starting ", *e)

	e.lds = createLDSServer(ldsPath)
	e.rds = createRDSServer(rdsPath, e.lds)
	e.rds.run()
	e.lds.run(e.rds)

	err := e.cmd.Start()
	if err != nil {
		log.WithError(err).Error("Envoy: Starting failed")
		return nil
	}
	log.Info("Envoy: Process started at pid ", e.cmd.Process.Pid)
	return e
}

// StopEnvoy kills the Envoy process started with StartEnvoy. The gRPC API streams are terminated
// first.
func (e *Envoy) StopEnvoy() {
	log.Info("Envoy: Stopping process ", e.cmd.Process.Pid)
	e.rds.stop()
	e.lds.stop()
	err := e.cmd.Process.Kill()
	if err != nil {
		log.WithError(err).Fatal("Envoy: Stopping failed")
	}
	e.cmd.Wait()
}

// AddListener adds a listener to a running Envoy proxy.
func (e *Envoy) AddListener(name string, port uint16, l7rules policy.L7DataMap, isIngress bool) {
	e.lds.addListener(name, port, l7rules, isIngress)
}

// UpdateListener changes to the L7 rules of an existing Envoy Listener.
func (e *Envoy) UpdateListener(name string, l7rules policy.L7DataMap) {
	e.lds.updateListener(name, l7rules)
}

// RemoveListener removes an existing Envoy Listener.
func (e *Envoy) RemoveListener(name string) {
	e.lds.removeListener(name)
}

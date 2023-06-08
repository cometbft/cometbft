package main

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	rpctypes "github.com/tendermint/tendermint/rpc/core/types"
	e2e "github.com/tendermint/tendermint/test/e2e/pkg"
)

// Perturbs a running testnet.
func Perturb(testnet *e2e.Testnet) error {
	for _, node := range testnet.Nodes {
		for _, perturbation := range node.Perturbations {
			_, err := PerturbNode(node, perturbation)
			if err != nil {
				return err
			}
			time.Sleep(3 * time.Second) // give network some time to recover between each
		}
	}
	return nil
}

// PerturbNode perturbs a node with a given perturbation, returning its status
// after recovering.
func PerturbNode(node *e2e.Node, perturbation e2e.Perturbation) (*rpctypes.ResultStatus, error) {
	testnet := node.Testnet
	out, err := execComposeOutput(testnet.Dir, "ps", "-q", node.Name)
	if err != nil {
		return nil, err
	}
	name := node.Name
	upgraded := false
	if len(out) == 0 {
		name = name + "_u"
		upgraded = true
		logger.Info("perturb node", "msg",
			log.NewLazySprintf("Node %v already upgraded, operating on alternate container %v",
				node.Name, name))
	}

	switch perturbation {
	case e2e.PerturbationDisconnect:
		logger.Info("perturb node", "msg", log.NewLazySprintf("Disconnecting node %v...", node.Name))
		if err := execDocker("network", "disconnect", testnet.Name+"_"+testnet.Name, name); err != nil {
			return nil, err
		}
		time.Sleep(10 * time.Second)
		if err := execDocker("network", "connect", testnet.Name+"_"+testnet.Name, name); err != nil {
			return nil, err
		}

	case e2e.PerturbationKill:
		logger.Info("perturb node", "msg", log.NewLazySprintf("Killing node %v...", node.Name))
		if err := execCompose(testnet.Dir, "kill", "-s", "SIGKILL", name); err != nil {
			return nil, err
		}
		if err := execCompose(testnet.Dir, "start", name); err != nil {
			return nil, err
		}

	case e2e.PerturbationPause:
		logger.Info("perturb node", "msg", log.NewLazySprintf("Pausing node %v...", node.Name))
		if err := execCompose(testnet.Dir, "pause", name); err != nil {
			return nil, err
		}
		time.Sleep(10 * time.Second)
		if err := execCompose(testnet.Dir, "unpause", name); err != nil {
			return nil, err
		}

	case e2e.PerturbationRestart:
		logger.Info("perturb node", "msg", log.NewLazySprintf("Restarting node %v...", node.Name))
		if err := execCompose(testnet.Dir, "restart", name); err != nil {
			return nil, err
		}

	case e2e.PerturbationUpgrade:
		oldV := node.Version
		newV := node.Testnet.UpgradeVersion
		if upgraded {
			return nil, fmt.Errorf("node %v can't be upgraded twice from version '%v' to version '%v'",
				node.Name, oldV, newV)
		}
		if oldV == newV {
			logger.Info("perturb node", "msg",
				log.NewLazySprintf("Skipping upgrade of node %v to version '%v'; versions are equal.",
					node.Name, newV))
			break
		}
		logger.Info("perturb node", "msg",
			log.NewLazySprintf("Upgrading node %v from version '%v' to version '%v'...",
				node.Name, oldV, newV))

		if err := execCompose(testnet.Dir, "stop", name); err != nil {
			return nil, err
		}
		time.Sleep(10 * time.Second)
		if err := execCompose(testnet.Dir, "up", "-d", name+"_u"); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unexpected perturbation %q", perturbation)
	}

	status, err := waitForNode(node, 0, 20*time.Second)
	if err != nil {
		return nil, err
	}
	logger.Info("perturb node",
		"msg",
		log.NewLazySprintf("Node %v recovered at height %v", node.Name, status.SyncInfo.LatestBlockHeight))
	return status, nil
}

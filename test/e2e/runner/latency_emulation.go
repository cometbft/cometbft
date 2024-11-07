package main

import (
	"fmt"
	"slices"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/test/e2e/pkg/infra"
)

// tcCommands generates the content of a shell script that includes a list of tc
// (traffic control) commands to emulate latency from the given node to all
// other nodes in the testnet.
func tcCommands(node *e2e.Node, infp infra.Provider) ([]string, error) {
	allZones, zoneMatrix, err := e2e.LoadZoneLatenciesMatrix()
	if err != nil {
		return nil, err
	}
	nodeZoneIndex := slices.Index(allZones, node.Zone)

	tcCmds := []string{
		"#!/bin/sh",
		"set -e",

		// Delete any existing qdisc on the root of the eth0 interface.
		"tc qdisc del dev eth0 root 2> /dev/null || true",

		// Add a new root qdisc of type HTB with a default class of 10.
		"tc qdisc add dev eth0 root handle 1: htb default 10",

		// Add a root class with identifier 1:1 and a rate limit of 1 gigabit per second.
		"tc class add dev eth0 parent 1: classid 1:1 htb rate 1gbit",

		// Add a default class under the root class with identifier 1:10 and a rate limit of 1 gigabit per second.
		"tc class add dev eth0 parent 1:1 classid 1:10 htb rate 1gbit",

		// Add an SFQ qdisc to the default class with handle 10: to manage traffic with fairness.
		"tc qdisc add dev eth0 parent 1:10 handle 10: sfq perturb 10",
	}

	// handle must be unique for each rule; start from one higher than last handle used above (10).
	handle := 11
	for _, targetZone := range allZones {
		// Get latency from node's zone to target zone (note that the matrix is symmetric).
		latency := zoneMatrix[targetZone][nodeZoneIndex]
		if latency <= 0 {
			continue
		}

		// Assign latency +/- 0.05% to handle.
		delta := latency / 20
		if delta == 0 {
			// Zero is not allowed in normal distribution.
			delta = 1
		}

		// Add a class with the calculated handle, under the root class, with the specified rate.
		tcCmds = append(tcCmds, fmt.Sprintf("tc class add dev eth0 parent 1:1 classid 1:%d htb rate 1gbit", handle))

		// Add a netem qdisc to simulate the specified delay with normal distribution.
		tcCmds = append(tcCmds, fmt.Sprintf("tc qdisc add dev eth0 parent 1:%d handle %d: netem delay %dms %dms distribution normal", handle, handle, latency, delta))

		// Set emulated latency to nodes in the target zone.
		for _, otherNode := range node.Testnet.Nodes {
			if otherNode.Zone == targetZone || node.Name == otherNode.Name {
				continue
			}
			otherNodeIP := infp.NodeIP(otherNode)
			// Assign latency handle to target node.
			tcCmds = append(tcCmds, fmt.Sprintf("tc filter add dev eth0 protocol ip parent 1: prio 1 u32 match ip dst %s/32 flowid 1:%d", otherNodeIP, handle))
		}

		handle++
	}

	// Display tc configuration for debugging.
	tcCmds = append(tcCmds, []string{
		fmt.Sprintf("echo Traffic Control configuration on %s:", node.Name),
		"tc qdisc show",
		"tc class show dev eth0",
		// "tc filter show dev eth0", // too verbose
	}...)

	return tcCmds, nil
}

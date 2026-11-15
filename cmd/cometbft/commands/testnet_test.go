package commands

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIncrementIP(t *testing.T) {
	{
		ip := net.ParseIP("192.168.1.10")
		assert.NotNil(t, ip)

		incrementIP(ip)
		expected := net.ParseIP("192.168.1.11")
		assert.True(t, ip.Equal(expected), "Expected %s to be equal to %s", ip, expected)

		incrementIP(ip)
		expected = net.ParseIP("192.168.1.12")
		assert.True(t, ip.Equal(expected), "Expected %s to be equal to %s", ip, expected)

		// Increment sufficiently to roll over
		for range 255 {
			incrementIP(ip)
		}
		expected = net.ParseIP("192.168.2.11")
		assert.True(t, ip.Equal(expected), "Expected %s to be equal to %s", ip, expected)
	}

	// Test case to test rollover of every triplet
	{
		ip := net.ParseIP("10.255.255.255")
		assert.NotNil(t, ip)

		incrementIP(ip)
		expected := net.ParseIP("11.0.0.0")
		assert.True(t, ip.Equal(expected), "Expected %s to be equal to %s", ip, expected)
	}

	// Test IPv6 for good measure.
	{
		ip := net.ParseIP("2a00:1450:400a:1009::65")
		assert.NotNil(t, ip)

		incrementIP(ip)
		expected := net.ParseIP("2a00:1450:400a:1009::66")
		assert.True(t, ip.Equal(expected), "Expected %s to be equal to %s", ip, expected)

		// Increment sufficiently to roll over
		for range 0xFFFF {
			incrementIP(ip)
		}
		expected = net.ParseIP("2a00:1450:400a:1009::1:65")
		assert.True(t, ip.Equal(expected), "Expected %s to be equal to %s", ip, expected)
	}
}

func TestTestnetFilesRejectsNegativeCounts(t *testing.T) {
	for name, tc := range map[string]struct{ v, n int }{
		"negative validators":     {v: -1, n: 0},
		"negative non-validators": {v: 4, n: -1},
	} {
		t.Run(name, func(t *testing.T) {
			oldV, oldN, oldOut, oldHosts, oldCfg := nValidators, nNonValidators, outputDir, hostnames, configFile
			t.Cleanup(func() {
				nValidators, nNonValidators, outputDir, hostnames, configFile = oldV, oldN, oldOut, oldHosts, oldCfg
			})
			nValidators, nNonValidators = tc.v, tc.n
			outputDir = t.TempDir()
			hostnames = nil
			configFile = ""

			require.NotPanics(t, func() {
				require.Error(t, testnetFiles(nil, nil))
			})
		})
	}
}

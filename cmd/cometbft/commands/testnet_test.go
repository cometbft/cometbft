package commands

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
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

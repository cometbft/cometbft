package netaddr

// func Test_String(t *testing.T) {
// 	tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
// 	require.NoError(t, err)

// 	netAddr := New("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", tcpAddr)

// 	var wg sync.WaitGroup

// 	for i := 0; i < 10; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			_ = netAddr.String()
// 		}()
// 	}

// 	wg.Wait()

// 	s := netAddr.String()
// 	require.Equal(t, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080", s)
// }

// func TestNew(t *testing.T) {
// 	tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
// 	require.NoError(t, err)

// 	assert.Panics(t, func() {
// 		New("", tcpAddr)
// 	})

// 	addr := New("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", tcpAddr)
// 	assert.Equal(t, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080", addr.String())

// 	assert.NotPanics(t, func() {
// 		New("", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8000})
// 	}, "Calling New with UDPAddr should not panic in testing")
// }

// func TestNewFromString(t *testing.T) {
// 	testCases := []struct {
// 		name     string
// 		addr     string
// 		expected string
// 		correct  bool
// 	}{
// 		{"no node id and no protocol", "127.0.0.1:8080", "", false},
// 		{"no node id w/ tcp input", "tcp://127.0.0.1:8080", "", false},
// 		{"no node id w/ udp input", "udp://127.0.0.1:8080", "", false},

// 		{
// 			"no protocol",
// 			"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080",
// 			"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080",
// 			true,
// 		},
// 		{
// 			"tcp input",
// 			"tcp://deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080",
// 			"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080",
// 			true,
// 		},
// 		{
// 			"udp input",
// 			"udp://deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080",
// 			"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080",
// 			true,
// 		},
// 		{"malformed tcp input", "tcp//deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080", "", false},
// 		{"malformed udp input", "udp//deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080", "", false},

// 		// {"127.0.0:8080", false},
// 		{"invalid host", "notahost", "", false},
// 		{"invalid port", "127.0.0.1:notapath", "", false},
// 		{"invalid host w/ port", "notahost:8080", "", false},
// 		{"just a port", "8082", "", false},
// 		{"non-existent port", "127.0.0:8080000", "", false},

// 		{"too short nodeId", "deadbeef@127.0.0.1:8080", "", false},
// 		{"too short, not hex nodeId", "this-isnot-hex@127.0.0.1:8080", "", false},
// 		{"not hex nodeId", "xxxxbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080", "", false},

// 		{"too short nodeId w/tcp", "tcp://deadbeef@127.0.0.1:8080", "", false},
// 		{"too short notHex nodeId w/tcp", "tcp://this-isnot-hex@127.0.0.1:8080", "", false},
// 		{"notHex nodeId w/tcp", "tcp://xxxxbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080", "", false},
// 		{
// 			"correct nodeId w/tcp",
// 			"tcp://deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080",
// 			"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080",
// 			true,
// 		},

// 		{"no node id", "tcp://@127.0.0.1:8080", "", false},
// 		{"no node id or IP", "tcp://@", "", false},
// 		{"tcp no host, w/ port", "tcp://:26656", "", false},
// 		{"empty", "", "", false},
// 		{"node id delimiter 1", "@", "", false},
// 		{"node id delimiter 2", " @", "", false},
// 		{"node id delimiter 3", " @ ", "", false},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			addr, err := NewFromString(tc.addr)
// 			if tc.correct {
// 				if assert.NoError(t, err, tc.addr) { //nolint:testifylint // require.Error doesn't work with the conditional here
// 					assert.Equal(t, tc.expected, addr.String())
// 				}
// 			} else {
// 				require.ErrorAs(t, err, &ErrInvalid{Addr: addr.String(), Err: err})
// 			}
// 		})
// 	}
// }

// func TestNewFromStrings(t *testing.T) {
// 	addrs, errs := NewFromStrings([]string{
// 		"127.0.0.1:8080",
// 		"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080",
// 		"deadbeefdeadbeefdeadbeefdeadbeefdeadbeed@127.0.0.2:8080",
// 	})
// 	assert.Len(t, addrs, 2)
// 	assert.Len(t, errs, 1)
// }

// func TestTCPFromIPPort(t *testing.T) {
// 	addr := TCPFromIPPort(net.ParseIP("127.0.0.1"), 8080)
// 	assert.Equal(t, "127.0.0.1:8080", addr.String())
// }

// func TestProperties(t *testing.T) {
// 	// TODO add more test cases
// 	testCases := []struct {
// 		addr     string
// 		valid    bool
// 		local    bool
// 		routable bool
// 	}{
// 		{"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@127.0.0.1:8080", true, true, false},
// 		{"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef@ya.ru:80", true, false, true},
// 	}

// 	for _, tc := range testCases {
// 		addr, err := NewFromString(tc.addr)
// 		require.NoError(t, err)

// 		err = addr.Valid()
// 		if tc.valid {
// 			require.NoError(t, err)
// 		} else {
// 			require.Error(t, err)
// 		}
// 		assert.Equal(t, tc.local, addr.Local())
// 		assert.Equal(t, tc.routable, addr.Routable())
// 	}
// }

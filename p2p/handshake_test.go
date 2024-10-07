package p2p

// func TestHandshake(t *testing.T) {
// 	ln, err := net.Listen("tcp", "127.0.0.1:0")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	var (
// 		peerPV       = ed25519.GenPrivKey()
// 		peerNodeInfo = testNodeInfo(key.PubKeyToID(peerPV.PubKey()), defaultNodeName)
// 	)

// 	go func() {
// 		c, err := net.Dial(ln.Addr().Network(), ln.Addr().String())
// 		if err != nil {
// 			t.Error(err)
// 			return
// 		}

// 		go func(c net.Conn) {
// 			_, err := protoio.NewDelimitedWriter(c).WriteMsg(peerNodeInfo.(DefaultNodeInfo).ToProto())
// 			if err != nil {
// 				t.Error(err)
// 			}
// 		}(c)
// 		go func(c net.Conn) {
// 			// ni   DefaultNodeInfo
// 			var pbni tmp2p.DefaultNodeInfo

// 			protoReader := protoio.NewDelimitedReader(c, ni.MaxNodeInfoSize())
// 			_, err := protoReader.ReadMsg(&pbni)
// 			if err != nil {
// 				t.Error(err)
// 			}

// 			_, err = ni.DefaultNodeInfoFromToProto(&pbni)
// 			if err != nil {
// 				t.Error(err)
// 			}
// 		}(c)
// 	}()

// 	_, err = ln.Accept()
// 	require.NoError(t, err)

// 	// ni, err := handshake(c, 20*time.Millisecond, emptyNodeInfo())
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }

// 	// if have, want := ni, peerNodeInfo; !reflect.DeepEqual(have, want) {
// 	// 	t.Errorf("have %v, want %v", have, want)
// 	// }
// }

/*
 *  Copyright (c) 2021 Neil Alexander
 *  Copyright (c) 2024 ruvcoindev
 *  This Source Code Form is subject to the terms of the Mozilla Public
 *  License, v. 2.0. If a copy of the MPL was not distributed with this
 *  file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package transport

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"regexp"
	"sync"
	"time"

	iwt "github.com/Arceliar/ironwood/types"
	"github.com/fatih/color"
	gologme "github.com/gologme/log"
	"github.com/quic-go/quic-go"
	"github.com/ruvcoindev/ruvchain-go/src/config"
	"github.com/ruvcoindev/ruvchain-go/src/core"
	"github.com/ruvcoindev/ruvchain-go/src/multicast"
)

type RuvchainTransport struct {
	listener   *quic.Listener
	ruvchain  net.PacketConn
	transport  *quic.Transport
	tlsConfig  *tls.Config
	quicConfig *quic.Config
	incoming   chan *ruvchainSession
	sessions   sync.Map // string -> quic.Connection
	dials      sync.Map // string -> *ruvchainDial
}

type ruvchainSession struct {
	quic.Connection
	quic.Stream
}

type ruvchainDial struct {
	context.Context
	context.CancelFunc
}

func NewRuvchainTransport(log *log.Logger, sk ed25519.PrivateKey, pk ed25519.PublicKey, peers []string, mcast bool) (*RuvchainTransport, error) {
	yellow := color.New(color.FgYellow).SprintfFunc()
	glog := gologme.New(log.Writer(), fmt.Sprintf("[ %s ] ", yellow("Ruvchain")), gologme.LstdFlags|gologme.Lmsgprefix)
	glog.EnableLevel("warn")
	glog.EnableLevel("error")
	glog.EnableLevel("info")

	cfg := config.GenerateConfig()
	copy(cfg.PrivateKey, sk)
	if err := cfg.GenerateSelfSignedCertificate(); err != nil {
		return nil, err
	}

	var ruv *core.Core
	var err error

	// Setup the Ruvchain node itself.
	{
		options := []core.SetupOption{
			core.NodeInfo(map[string]interface{}{
				"name": hex.EncodeToString(pk) + "@ruvmail",
			}),
			core.NodeInfoPrivacy(true),
		}
		for _, peer := range peers {
			options = append(options, core.Peer{URI: peer})
		}
		if ygg, err = core.New(cfg.Certificate, glog, options...); err != nil {
			panic(err)
		}
	}

	// Setup the multicast module.
	{
		options := []multicast.SetupOption{
			multicast.MulticastInterface{
				Regex:  regexp.MustCompile(".*"),
				Beacon: mcast,
				Listen: mcast,
			},
		}
		if _, err = multicast.New(ruv, glog, options...); err != nil {
			panic(err)
		}
	}

	tr := &RuvchainTransport{
		tlsConfig: &tls.Config{
			ServerName: hex.EncodeToString(ruv.PublicKey()),
			Certificates: []tls.Certificate{
				*cfg.Certificate,
			},
			InsecureSkipVerify: true,
		},
		quicConfig: &quic.Config{
			HandshakeIdleTimeout: time.Second * 5,
			MaxIdleTimeout:       time.Second * 60,
		},
		transport: &quic.Transport{
			Conn: ruv,
		},
		ruvchain: ruv,
		incoming:  make(chan *ruvchainSession, 1),
	}

	if tr.listener, err = tr.transport.Listen(tr.tlsConfig, tr.quicConfig); err != nil {
		return nil, fmt.Errorf("quic.Listen: %w", err)
	}

	go tr.connectionAcceptLoop()
	return tr, nil
}

func (t *RuvchainTransport) connectionAcceptLoop() {
	for {
		qc, err := t.listener.Accept(context.TODO())
		if err != nil {
			return
		}

		host := qc.RemoteAddr().String()
		if eqc, ok := t.sessions.LoadAndDelete(host); ok {
			eqc := eqc.(quic.Connection)
			_ = eqc.CloseWithError(0, "Connection replaced")
		}
		t.sessions.Store(host, qc)
		if dial, ok := t.dials.LoadAndDelete(host); ok {
			dial := dial.(*ruvchainDial)
			dial.CancelFunc()
		}

		go t.streamAcceptLoop(qc)
	}
}

func (t *RuvchainTransport) streamAcceptLoop(qc quic.Connection) {
	host := qc.RemoteAddr().String()

	defer qc.CloseWithError(0, "Timed out") // nolint:errcheck
	defer t.sessions.Delete(host)

	for {
		qs, err := qc.AcceptStream(context.Background())
		if err != nil {
			break
		}
		t.incoming <- &ruvchainSession{qc, qs}
	}
}

func (t *RuvchainTransport) Dial(host string) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	var retry bool
retry:
	qc, ok := t.sessions.Load(host)
	if !ok {
		if dial, ok := t.dials.Load(host); ok {
			<-dial.(*ruvchainDial).Done()
		}
		if qc, ok = t.sessions.Load(host); !ok {
			dialctx, dialcancel := context.WithCancel(ctx)
			defer dialcancel()

			t.dials.Store(host, &ruvchainDial{dialctx, dialcancel})
			defer t.dials.Delete(host)

			addr := make(iwt.Addr, ed25519.PublicKeySize)
			k, err := hex.DecodeString(host)
			if err != nil {
				return nil, err
			}
			copy(addr, k)

			if qc, err = t.transport.Dial(dialctx, addr, t.tlsConfig, t.quicConfig); err != nil {
				return nil, err
			}

			qc := qc.(quic.Connection)
			t.sessions.Store(host, qc)
			go t.streamAcceptLoop(qc)
		}
	}
	if qc == nil {
		return nil, net.ErrClosed
	} else {
		qc := qc.(quic.Connection)
		qs, err := qc.OpenStreamSync(ctx)
		if err != nil {
			if !retry {
				retry = true
				goto retry
			}
			return nil, err
		}
		// For some reason this is needed to kick the stream
		_, err = qs.Write([]byte(" "))
		return &ruvchainSession{qc, qs}, err
	}
}

func (t *RuvchainTransport) Listener() net.Listener {
	return &ruvchainListener{t}
}

type ruvchainListener struct {
	*RuvchainTransport
}

func (t *ruvchainListener) Accept() (net.Conn, error) {
	return <-t.incoming, nil
}

func (t *ruvchainListener) Addr() net.Addr {
	return t.listener.Addr()
}

func (t *ruvchainListener) Close() error {
	if err := t.listener.Close(); err != nil {
		return err
	}
	return t.ruvchain.Close()
}

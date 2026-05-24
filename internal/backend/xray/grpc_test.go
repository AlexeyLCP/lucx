//go:build integration
// +build integration

package xray

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/xtls/xray-core/app/proxyman"
	proxymanCmd "github.com/xtls/xray-core/app/proxyman/command"
	"github.com/xtls/xray-core/app/router"
	routerCmd "github.com/xtls/xray-core/app/router/command"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/proxy/vless"
	vlessInbound "github.com/xtls/xray-core/proxy/vless/inbound"
	vlessOutbound "github.com/xtls/xray-core/proxy/vless/outbound"
	"github.com/xtls/xray-core/transport/internet"
	"github.com/xtls/xray-core/transport/internet/reality"
)

func getGRPCAddr() string {
	if addr := os.Getenv("XRAY_GRPC_ADDR"); addr != "" {
		return addr
	}
	return "localhost:10085"
}

func connectGRPC(t *testing.T) (*grpc.ClientConn, context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	addr := getGRPCAddr()
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		t.Skipf("Cannot connect to Xray gRPC at %s: %v -- skipping integration test", addr, err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn, ctx
}

func singlePort(p uint32) *net.PortList {
	return &net.PortList{Range: []*net.PortRange{{From: p, To: p}}}
}

func vlessUser(id, flow string) *protocol.User {
	return &protocol.User{
		Account: serial.ToTypedMessage(&vless.Account{
			Id:   id,
			Flow: flow,
		}),
	}
}

// TestGRPCAddRemoveInbound tests the gRPC HandlerService for VLESS inbound
// add/remove operations.
func TestGRPCAddRemoveInbound(t *testing.T) {
	conn, ctx := connectGRPC(t)
	client := proxymanCmd.NewHandlerServiceClient(conn)

	t.Log("=== Test 1: AddInbound (VLESS) ===")
	_, err := client.AddInbound(ctx, &proxymanCmd.AddInboundRequest{
		Inbound: &core.InboundHandlerConfig{
			Tag: "lucx-test-inbound",
			ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
				PortList: singlePort(12345),
			}),
			ProxySettings: serial.ToTypedMessage(&vlessInbound.Config{
				Clients:    []*protocol.User{vlessUser("test-uuid-0000-0000-000000000000", "")},
				Decryption: "none",
			}),
		},
	})
	if err != nil {
		t.Logf("AddInbound FAILED: %v", err)
		t.Log("-> gRPC AddInbound not supported or params rejected")
		t.Log("-> Fallback to config.json required for inbound configs")
		return
	}
	t.Log("AddInbound OK")

	t.Log("=== Test 2: RemoveInbound ===")
	_, err = client.RemoveInbound(ctx, &proxymanCmd.RemoveInboundRequest{
		Tag: "lucx-test-inbound",
	})
	if err != nil {
		t.Logf("RemoveInbound FAILED: %v", err)
		return
	}
	t.Log("RemoveInbound OK")

	t.Log("=== RESULT: gRPC HandlerService works for AddInbound/RemoveInbound ===")
}

// TestGRPCAddRemoveOutbound tests the gRPC HandlerService for VLESS outbound
// add/remove operations.
func TestGRPCAddRemoveOutbound(t *testing.T) {
	conn, ctx := connectGRPC(t)
	client := proxymanCmd.NewHandlerServiceClient(conn)

	t.Log("=== Test 3: AddOutbound (VLESS) ===")
	_, err := client.AddOutbound(ctx, &proxymanCmd.AddOutboundRequest{
		Outbound: &core.OutboundHandlerConfig{
			Tag: "lucx-test-outbound",
			SenderSettings: serial.ToTypedMessage(&proxyman.SenderConfig{}),
			ProxySettings: serial.ToTypedMessage(&vlessOutbound.Config{
				Vnext: &protocol.ServerEndpoint{
					Address: &net.IPOrDomain{
						Address: &net.IPOrDomain_Domain{Domain: "1.2.3.4"},
					},
					Port: 443,
					User: vlessUser("test-uuid-outbound", ""),
				},
			}),
		},
	})
	if err != nil {
		t.Logf("AddOutbound FAILED: %v", err)
		t.Log("-> gRPC AddOutbound not supported")
		return
	}
	t.Log("AddOutbound OK")

	t.Log("=== Test 4: RemoveOutbound ===")
	_, err = client.RemoveOutbound(ctx, &proxymanCmd.RemoveOutboundRequest{
		Tag: "lucx-test-outbound",
	})
	if err != nil {
		t.Logf("RemoveOutbound FAILED: %v", err)
		return
	}
	t.Log("RemoveOutbound OK")

	t.Log("=== RESULT: gRPC HandlerService works for AddOutbound/RemoveOutbound ===")
}

// TestGRPCRoutingService tests the gRPC RoutingService for rule add/remove.
func TestGRPCRoutingService(t *testing.T) {
	conn, ctx := connectGRPC(t)
	routingClient := routerCmd.NewRoutingServiceClient(conn)

	t.Log("=== Test 5: RoutingService.AddRule ===")
	_, err := routingClient.AddRule(ctx, &routerCmd.AddRuleRequest{
		Config: serial.ToTypedMessage(&router.RoutingRule{
			TargetTag: &router.RoutingRule_Tag{Tag: "test-out"},
			RuleTag:   "lucx-test-rule",
		}),
		ShouldAppend: true,
	})
	if err != nil {
		t.Logf("RoutingService.AddRule FAILED: %v", err)
		t.Log("-> RoutingService not available. Routing must be set via config file.")
		return
	}
	t.Log("RoutingService.AddRule OK")

	t.Log("=== Test 6: RoutingService.RemoveRule ===")
	_, err = routingClient.RemoveRule(ctx, &routerCmd.RemoveRuleRequest{
		RuleTag: "lucx-test-rule",
	})
	if err != nil {
		t.Logf("RoutingService.RemoveRule FAILED: %v", err)
		return
	}
	t.Log("RoutingService.RemoveRule OK")

	t.Log("=== RESULT: gRPC RoutingService works ===")
}

// TestGRPCRealityUTLSFullConfig tests whether gRPC HandlerService can handle
// the full Reality + uTLS configuration (the most complex config LucX needs).
func TestGRPCRealityUTLSFullConfig(t *testing.T) {
	conn, ctx := connectGRPC(t)
	client := proxymanCmd.NewHandlerServiceClient(conn)

	// Decode known test key material.
	privKey, _ := base64.StdEncoding.DecodeString("2KZ2uHSVFqqWSnB3BBc7YKgLD4BJF5BNBDhanoWdHhc")
	pubKey, _ := base64.StdEncoding.DecodeString("GtPQkT+ZQIF2JROrWM5CmhB6FWYZkLe+QXJFpDWhGFg")
	shortID, _ := hex.DecodeString("6ba85179")

	t.Log("=== Test 7: Full Reality+uTLS+XTLS config ===")
	_, err := client.AddInbound(ctx, &proxymanCmd.AddInboundRequest{
		Inbound: &core.InboundHandlerConfig{
			Tag: "lucx-test-reality-full",
			ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
				PortList: singlePort(12346),
				StreamSettings: &internet.StreamConfig{
					ProtocolName: "tcp",
					SecurityType: "reality",
					SecuritySettings: []*serial.TypedMessage{
						serial.ToTypedMessage(&reality.Config{
							ServerNames: []string{"discord.com", "apple.com", "cloudflare.com"},
							PrivateKey:  privKey,
							ShortIds:    [][]byte{shortID},
							PublicKey:   pubKey,
							Fingerprint: "chrome",
							Show:        true,
							Dest:        "discord.com:443",
						}),
					},
				},
			}),
			ProxySettings: serial.ToTypedMessage(&vlessInbound.Config{
				Clients:    []*protocol.User{vlessUser("full-test-uuid-0000-0000-000000000000", "xtls-rprx-vision")},
				Decryption: "none",
			}),
		},
	})
	if err != nil {
		t.Logf("Full Reality config FAILED: %v", err)
		t.Log("-> CONCLUSION: gRPC cannot handle complex Reality configs")
		t.Log("-> RECOMMENDATION: Use config.json for Reality, gRPC for simple configs")
		return
	}
	t.Log("Full Reality config OK -- gRPC handles complex Reality!")
	client.RemoveInbound(ctx, &proxymanCmd.RemoveInboundRequest{Tag: "lucx-test-reality-full"})
}

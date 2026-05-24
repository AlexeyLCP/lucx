# Phase 0: Xray gRPC API Verification

## Prerequisites
- Running Xray 1.8.x+ with gRPC HandlerService enabled on port 10085
- `grpcurl` or Go installed

## Setup Xray for gRPC

Add to Xray config.json:
```json
{
  "api": {
    "services": ["HandlerService", "RoutingService"],
    "tag": "api"
  },
  "inbounds": [{
    "tag": "api",
    "listen": "127.0.0.1",
    "port": 10085,
    "protocol": "dokodemo-door",
    "settings": { "address": "127.0.0.1" }
  }],
  "policy": {
    "system": {
      "statsInboundUplink": true,
      "statsInboundDownlink": true
    }
  },
  "routing": {
    "rules": [{
      "type": "field",
      "inboundTag": ["api"],
      "outboundTag": "api",
      "targetTag": "api"
    }]
  }
}
```

Restart Xray: `systemctl restart xray`

## Quick Test with grpcurl

```bash
# List services
grpcurl -plaintext localhost:10085 list

# Test AddInbound (simple VLESS)
grpcurl -plaintext -d '{"inbound":{"tag":"test","receiver_settings":{"@type":"xray.app.proxyman.ReceiverConfig","port_list":{"range":[{"From":12345,"To":12345}]}},"proxy_settings":{"@type":"xray.proxy.vless.inbound.Config","clients":[{"account":{"@type":"xray.proxy.vless.Account","id":"test-uuid-0000-0000-000000000000"}}],"decryption":"none"}}}' \
  localhost:10085 xray.app.proxyman.command.HandlerService/AddInbound
```

## Run Go Integration Tests

```bash
export XRAY_GRPC_ADDR=localhost:10085
go test -v -run TestGRPC -tags=integration ./internal/backend/xray/
```

## Decision Matrix

| Test | Result | Action |
|------|--------|--------|
| AddInbound works | Yes | Use gRPC for inbounds |
| AddInbound fails | No | Use config.json fallback |
| Reality+uTLS works | Yes | gRPC for all configs |
| Reality+uTLS fails | No | Hybrid: simple via gRPC, Reality via config |
| RoutingService works | Yes | gRPC routing |
| RoutingService fails | No | Config file routing |

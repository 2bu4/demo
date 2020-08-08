package userip

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

func FromRequest(req *http.Request) (net.IP, error) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}
	return userIP, nil
}

type key int

const userIPKey key = 0

func NewContext(ctx context.Context, userIp net.IP) context.Context {
	return context.WithValue(ctx, userIPKey, userIp)
}

func FromContext(ctx context.Context) (net.IP, bool) {
	userIp, ok := ctx.Value(userIPKey).(net.IP)
	return userIp, ok
}

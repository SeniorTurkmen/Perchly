package requestlog

import "net/http"

// Device header contract: every client request is expected to carry
// these (see README.md). All are optional from the server's point of
// view — a missing header just logs as null, never rejected.
const (
	HeaderClientVersion = "X-Client-Version" // e.g. "1.4.2"
	HeaderPlatform      = "X-Platform"       // e.g. "ios", "android"
	HeaderOSVersion     = "X-OS-Version"     // e.g. "26.1"
	HeaderDeviceModel   = "X-Device-Model"   // e.g. "iPhone17,2"
)

type deviceInfo struct {
	clientVersion *string
	platform      *string
	osVersion     *string
	deviceModel   *string
}

func parseDeviceInfo(r *http.Request) deviceInfo {
	return deviceInfo{
		clientVersion: headerPtr(r, HeaderClientVersion),
		platform:      headerPtr(r, HeaderPlatform),
		osVersion:     headerPtr(r, HeaderOSVersion),
		deviceModel:   headerPtr(r, HeaderDeviceModel),
	}
}

func headerPtr(r *http.Request, key string) *string {
	v := r.Header.Get(key)
	if v == "" {
		return nil
	}
	return &v
}

package iplist

import (
	"net/netip"
)

// IPList represents a single normalized parsed entry from your data source.
type IPList struct {
	IP     string
	Enable bool
}

// New instantiates a normalized IPList entry.
func New(ipStr string, enable int) IPList {
	return IPList{
		IP:     ipStr,
		Enable: enable == 1,
	}
}

// NewEnabled creates an enabled IPList entry.
func NewEnabled(ip string) IPList {
	return IPList{
		IP:     ip,
		Enable: true,
	}
}

// NewDisabled creates a disabled IPList entry.
func NewDisabled(ip string) IPList {
	return IPList{
		IP:     ip,
		Enable: false,
	}
}

// EncodeCSV returns the CSV representation of the entry.
//
// Format:
//
//	ip,enable
//
// Example:
//
//	192.168.1.1,1
//	10.0.0.0/24,0
func (e IPList) EncodeCSV() []string {
	enable := "0"
	if e.Enable {
		enable = "1"
	}

	return []string{e.IP, enable}
}

// IsCIDR returns true if the entry represents a subnet range.
func (i IPList) IsCIDR() bool {
	return len(i.IP) > 0 && (i.IP[len(i.IP)-1] >= '0' && i.IP[len(i.IP)-1] <= '9') && (i.IP[0] >= '0' && i.IP[0] <= '9') && (i.IP[len(i.IP)-2] == '/' || i.IP[len(i.IP)-3] == '/')
}

// CIDRBlock stores the mathematical boundaries for subnets.
type CIDRBlock struct {
	StartIP   netip.Addr
	TotalIPs  uint64
	GlobalIdx uint64
}

// MasterIndexer tracks the hybrid data schema across files.
type MasterIndexer struct {
	FilePath      string
	CIDRBlocks    []CIDRBlock
	SingleOffsets []int64
	TotalCIDRIPs  uint64
	TotalSingles  uint64
	GrandTotal    uint64
}

package vm

import (
	"context"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"time"

	goipam "github.com/metal-stack/go-ipam"
	"github.com/vishvananda/netlink"
)

type Ipam struct {
	ipamer goipam.Ipamer
	prefix *goipam.Prefix
}

const (
	BridgeName = "pwbr0"
)

func NewIpam(subnet string) (*Ipam, error) {
	ipam := goipam.New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	prefix, err := ipam.NewPrefix(ctx, subnet)
	if err != nil {
		return nil, err
	}
	return &Ipam{ipamer: ipam, prefix: prefix}, nil
}

func (ipam *Ipam) AcquireIP() (netip.Addr, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ip, err := ipam.ipamer.AcquireIP(ctx, ipam.prefix.Cidr)
	if err != nil {
		return netip.Addr{}, err
	}
	return ip.IP, err
}

func (ipam *Ipam) AcquireSpecificIP(ipAddress string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := ipam.ipamer.AcquireSpecificIP(ctx, ipam.prefix.Cidr, ipAddress)
	return err
}

func (ipam *Ipam) ReleaseIP(ipAddress string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := ipam.ipamer.ReleaseIPFromPrefix(ctx, ipam.prefix.Cidr, ipAddress)
	return err
}

func (ipam *Ipam) GetCIDR() string {
	return ipam.prefix.Cidr
}

func getNextTap() (string, error) {
	linkList, err := netlink.LinkList()
	if err != nil {
		return "", err
	}
	largestTap := -1
	for _, link := range linkList {
		if link.Type() == "tuntap" {
			name := link.Attrs().Name
			vals := strings.Split(name, "tap")
			tapNum, err := strconv.Atoi(vals[1])
			if err != nil {
				return "", err
			}
			if tapNum > largestTap {
				largestTap = tapNum
			}
		}
	}
	newTapNum := largestTap + 1
	newTap := fmt.Sprintf("tap%d", newTapNum)
	la := netlink.NewLinkAttrs()
	la.Name = newTap
	tap := &netlink.Tuntap{LinkAttrs: la, Mode: netlink.TUNTAP_MODE_TAP}
	err = netlink.LinkAdd(tap)
	if err != nil {
		return "", err
	}
	bridge, err := netlink.LinkByName(BridgeName)
	if err != nil {
		return "", err
	}
	err = netlink.LinkSetMaster(tap, bridge)
	if err != nil {
		return "", err
	}
	err = netlink.LinkSetUp(tap)
	return newTap, err
}

func releaseTap(tapName string) error {
	la := netlink.NewLinkAttrs()
	la.Name = tapName
	tap := &netlink.Tuntap{LinkAttrs: la, Mode: netlink.TUNTAP_MODE_TAP}
	err := netlink.LinkSetDown(tap)
	if err != nil {
		return err
	}
	err = netlink.LinkDel(tap)
	if err != nil {
		return err
	}
	return nil
}

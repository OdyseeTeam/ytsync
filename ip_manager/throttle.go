package ip_manager

import (
	"net"
	"sort"
	"sync"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/lbryio/lbry.go/v2/extras/errors"
	"github.com/lbryio/lbry.go/v2/extras/stop"
	"github.com/lbryio/ytsync/v5/util"
	log "github.com/sirupsen/logrus"
)

const IPCooldownPeriod = 20 * time.Second
const unbanTimeout = 48 * time.Hour

var stopper = stop.New()

type IPPool struct {
	ips     []throttledIP
	lock    *sync.RWMutex
	stopGrp *stop.Group
}

type throttledIP struct {
	IP           string
	UsedForVideo string
	LastUse      time.Time
	Throttled    bool
	InUse        bool
}

var ipPoolInstance *IPPool

func getIps() ([]throttledIP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, errors.Err(err)
	}
	var pool []throttledIP
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
			if ipnet.IP.To16() != nil && govalidator.IsIPv6(ipnet.IP.String()) {
				pool = append(pool, throttledIP{
					IP:      ipnet.IP.String(),
					LastUse: time.Now().Add(-5 * time.Minute),
				})
			} else if ipnet.IP.To4() != nil && govalidator.IsIPv4(ipnet.IP.String()) {
				pool = append(pool, throttledIP{
					IP:      ipnet.IP.String(),
					LastUse: time.Now().Add(-5 * time.Minute),
				})
			}
		}
	}
	return pool, nil
}

func GetIPPool(stopGrp *stop.Group) (*IPPool, error) {
	if ipPoolInstance != nil {
		return ipPoolInstance, nil
	}
	pool, err := getIps()
	if err != nil {
		return nil, err
	}
	ipPoolInstance = &IPPool{
		ips:     pool,
		lock:    &sync.RWMutex{},
		stopGrp: stopGrp,
	}
	return ipPoolInstance, nil
}

func (i *IPPool) UpdateIps() error {
	currentIPs, err := getIps()
	if err != nil {
		return err
	}
	newIPsMap := make(map[string]bool)
	for _, ip := range currentIPs {
		newIPsMap[ip.IP] = true
	}

	oldIpsMap := make(map[string]bool)
	refreshedIPs := make([]throttledIP, 0)
	for _, ip := range i.ips {
		oldIpsMap[ip.IP] = true
		if newIPsMap[ip.IP] {
			refreshedIPs = append(refreshedIPs, ip)
		}
	}

	for _, ip := range currentIPs {
		if !oldIpsMap[ip.IP] {
			refreshedIPs = append(refreshedIPs, ip)
		}
	}
	i.ips = refreshedIPs
	return nil
}

// AllThrottled checks whether the IPs provided are all throttled.
// returns false if at least one IP is not throttled
// Not thread safe, should use locking when called
func AllThrottled(ips []throttledIP) bool {
	for _, i := range ips {
		if !i.Throttled {
			return false
		}
	}
	return true
}

// AllInUse checks whether the IPs provided are all currently in use.
// returns false if at least one IP is not in use AND is not throttled
// Not thread safe, should use locking when called
func AllInUse(ips []throttledIP) bool {
	for _, i := range ips {
		if !i.InUse && !i.Throttled {
			return false
		}
	}
	return true
}

func (i *IPPool) ReleaseIP(ip string) {
	i.lock.Lock()
	defer i.lock.Unlock()
	for j := range i.ips {
		localIP := &i.ips[j]
		if localIP.IP == ip {
			localIP.InUse = false
			localIP.LastUse = time.Now()
			return
		}
	}
	util.SendErrorToSlack("something went wrong while releasing the IP %s as we reached the end of the function", ip)
}

func (i *IPPool) ReleaseAll() {
	i.lock.Lock()
	defer i.lock.Unlock()
	for j := range i.ips {
		if i.ips[j].Throttled {
			continue
		}
		localIP := &i.ips[j]
		localIP.InUse = false
	}
}

// SetThrottled sets the throttled flag for the provided IP and schedules its unban for a future time
// todo: this might introduce a leak if the ip is removed from the pool while it's throttled (this would be for the VPN interface)
func (i *IPPool) SetThrottled(ip string) {
	i.lock.Lock()
	defer i.lock.Unlock()
	var tIP *throttledIP
	for j, _ := range i.ips {
		localIP := &i.ips[j]
		if localIP.IP == ip {
			if localIP.Throttled {
				return
			}
			localIP.Throttled = true
			tIP = localIP
			break
		}
	}
	util.SendErrorToSlack("%s set to throttled", ip)

	stopper.Add(1)
	go func(tIP *throttledIP) {
		defer stopper.Done()
		unbanTimer := time.NewTimer(unbanTimeout)
		select {
		case <-unbanTimer.C:
			i.lock.Lock()
			tIP.Throttled = false
			i.lock.Unlock()
			util.SendInfoToSlack("%s set back to not throttled", ip)
		case <-i.stopGrp.Ch():
			unbanTimer.Stop()
		}
	}(tIP)
}

var ErrAllInUse = errors.Base("all IPs are in use, try again")
var ErrAllThrottled = errors.Base("all IPs are throttled")
var ErrResourceLock = errors.Base("error getting next ip, did you forget to lock on the resource?")
var ErrInterruptedByUser = errors.Base("interrupted by user")

func (i *IPPool) nextIP(forVideo string) (*throttledIP, error) {
	if i == nil {
		util.SendErrorToSlack("ip pool is nil")
	}
	if i.lock == nil {
		util.SendErrorToSlack("ip pool lock is nil")
	}
	i.lock.Lock()
	defer i.lock.Unlock()

	sort.Slice(i.ips, func(j, k int) bool {
		return i.ips[j].LastUse.Before(i.ips[k].LastUse)
	})

	if !AllThrottled(i.ips) {
		if AllInUse(i.ips) {
			return nil, errors.Err(ErrAllInUse)
		}

		var nextIP *throttledIP
		for j := range i.ips {
			ip := &i.ips[j]
			if ip.InUse || ip.Throttled {
				continue
			}
			nextIP = ip
			break
		}
		if nextIP == nil {
			return nil, errors.Err(ErrResourceLock)
		}
		nextIP.InUse = true
		nextIP.UsedForVideo = forVideo
		return nextIP, nil
	}
	return nil, errors.Err(ErrAllThrottled)
}

func (i *IPPool) GetIP(forVideo string) (string, error) {
	for {
		ip, err := i.nextIP(forVideo)
		if err != nil {
			if errors.Is(err, ErrAllInUse) {
				select {
				case <-i.stopGrp.Ch():
					return "", errors.Err(ErrInterruptedByUser)
				default:
					time.Sleep(5 * time.Second)
					continue
				}
			} else if errors.Is(err, ErrAllThrottled) {
				return "throttled", err
			}
			return "", err
		}
		if time.Since(ip.LastUse) < IPCooldownPeriod {
			log.Debugf("The IP %s is too hot, waiting for %.1f seconds before continuing", ip.IP, (IPCooldownPeriod - time.Since(ip.LastUse)).Seconds())
			time.Sleep(IPCooldownPeriod - time.Since(ip.LastUse))
		}
		return ip.IP, nil
	}
}

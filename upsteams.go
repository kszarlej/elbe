package main

import (
	"time"
)

var DynamicUpstreams map[string]Upstream

type LoadBalancer struct {
	last int
	max  int
	loop *[]string
}

// GetUpstream returns the next upstream from list of upstreams
// in a LoadBalancer. Simple RoundRobin.
func (lb *LoadBalancer) GetHost() string {
	hostNum := lb.last + 1

	if hostNum == lb.max {
		hostNum = 0
	}

	lb.last = hostNum

	return (*lb.loop)[hostNum]
}

// SetDynamicUpstreams is function which reads upstreams from config file
// and sets DynamicUpstreams struct used by loadbalancer.
func SetDynamicUpstreams(config *Config, init bool) {
	for {
		var upstreams = make(map[string]Upstream)

		for upstreamName, upstreamConfig := range config.Upstreams {
			var hosts []string
			var ups = Upstream{}

			for _, host := range upstreamConfig.Hosts {
				hosts = append(hosts, host)
			}

			ups.Hosts = hosts

			if DynamicUpstreams[upstreamName].LoadBalancer == nil {
				var loadbalancer = LoadBalancer{last: 0, max: len(hosts), loop: &ups.Hosts}
				ups.LoadBalancer = &loadbalancer // why this work
				//DynamicUpstreams[upstreamName].LoadBalancer = &loadbalancer // and this doesn't?
			} else {
				ups.LoadBalancer = DynamicUpstreams[upstreamName].LoadBalancer
				ups.LoadBalancer.max = len(hosts)
			}

			upstreams[upstreamName] = ups
		}

		// Set new upstreams
		DynamicUpstreams = upstreams

		if init == true {
			return
		}
		time.Sleep(5 * time.Second)
	}
}

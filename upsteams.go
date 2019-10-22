package main

import (
	"time"
)

var DynamicUpstreams map[string]Upstream

type roundrobin struct {
	last int
	max  int
	loop *[]string
}

// Function which reads upstreams from config file
// and sets DynamicUpstreams struct used by loadbalancer.
func SetDynamicUpstreams(config *Config, init bool) {
	//configUpstreams := config.Upstreams

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
				var loadbalancer = roundrobin{last: 0, max: len(hosts), loop: &ups.Hosts}
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

func RoundRobinGetHost(upstreamName string) string {
	ups := DynamicUpstreams[upstreamName]

	hostNum := ups.LoadBalancer.last + 1

	if hostNum == ups.LoadBalancer.max {
		hostNum = 0
	}

	ups.LoadBalancer.last = hostNum

	return ups.Hosts[hostNum]
}

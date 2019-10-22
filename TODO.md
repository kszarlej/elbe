# Version 0.2:

1. Properly handle Connection header: https://tools.ietf.org/html/rfc2616#section-14.10


# Version 0.1: 

1. Basic location
    * Prefixed location choice - *DONE*
2. Upstreams:
    * Automatic upstreams reloading from config
    * Simple RoundRobin loadbalancing across all Upstreams  *DONE*
3. Full HTTP/1.0 and HTTP/1.1 support:
    * Support for POST, HEAD, GET, POST, DELETE, OPTIONS methods
4. YAML Configuration
6. Support for Basic Authorization
7. Proxy module
    * Support for proxy_add_header - *DONE*
    * Support for proxy_hider_header - *DONE*
    * Support for proxy_read_timeout - *DONE*
    * Support for proxy_send_timeout - *DONE*
    * Support for proxy_set_body - *DONE*
    * Support for proxy_next_upstream 


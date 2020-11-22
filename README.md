# Dispatcher
### Overview
A high-performance request dispatcher, authentication checker and rate limit checker.  

**All the things Dispatcher does :**
1. Reads all route configs and inits a mux router on them
2. Inits redis connection pool
3. Starts listening for incoming requests
  
**Every incoming request :**
1. should be check in terms of rate limitation and ban condition
2. should be check in terms of authentication
3. will be proxy to the related service

### Ports and configs
- Defaults:
    * `:80` for `DEFAULT_ADDR`
    * `:6379` for `REDIS_ADDR`
    * `""` for `REDIS_PASS`
    * `""` for `HTTP_PREFIX`
    * `/app/route_configs.yaml` for `CONFIG_PATH` on docker  

- ENV vars:
    * `DEFAULT_ADDR` for Dispatcher
    * `REDIS_ADDR` redis server address
    * `REDIS_PASS` redis password
    * `HTTP_PREFIX` http prefix (if needed)
    * `CONFIG_PATH` dispatcher rule path  

- Make commands:
    * `make deps` get dependencies
    * `make build` build locally
    * `make test` run tests
    * `make clean` clean binaries
    * `make docker-build` build and run docker

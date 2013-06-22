Moln
---

Moln is eventually going to be a sync server similar to OwnCloud. I plan to make the server only available through APIs. Eventually I'd like to make clients the various popular platforms.

### Features
- [] User accounts
- [] User containers
- [] User access logs
- [] FS data sync
- [] Calendar sync
- [] Address book sync
- [] Task manager w/ sync?

### Running the server
Currently you'll have to start things manually, I'll add a Procfile later for foreman support.

1. Install Redis, and get the code with `go get`
2. Start Redis, with the environment config in `/config/redis/`
3. Start the server, if you used `go get` simply run `moln`, otherwise build and run the binary.

If you want to change the server environment set the `ENVIRONMENT` env variable(e.g. `ENVIRONMENT=production ./moln`).

### License
MIT, view the included [LICENSE](https://raw.github.com/larzconwell/moln/master/LICENSE) file for details.

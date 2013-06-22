Moln
---

Moln is eventually going to be a sync server similar to OwnCloud. I plan to make the server only available through APIs. Eventually I'd like to make clients the various popular platforms.

### Features
- [] Hypermedia APIs
- [] User accounts
- [] User containers
- [] User access logs
- [] FS data sync
- [] Calendar sync
- [] Address book sync
- [] Task manager w/ sync

### Running the server
1. Install Redis, install a [Foreman](https://github.com/ddollar/foreman) tool, and get the code with `go get`
2. Create any directories needed, `/var/log/redis`, `./db`
3. Start the foreman process(e.g. `foreman start -e config/development.env`) with your environment env file.

Note: When running in production mode on Windows, Redis connects to a UNIX Socket so you may encounter errors

### License
MIT, view the included [LICENSE](https://raw.github.com/larzconwell/moln/master/LICENSE) file for details.

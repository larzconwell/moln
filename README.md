Moln
---

Moln is a simple sync API server with features similar to OwnCloud, including multiple user accounts, file sync, and much more.

### Features
- [DONE] Multiple user accounts
- [DONE] Account access from devices
- [] Account access logs
- [] File sync
- [] Calendar sync
- [] Address book sync
- [] Task manager

### Running and Deploying

#### Prerequisites
1. Install Redis.
2. Go(only the development machine needs Go).
3. The [Foreman](https://github.com/ddollar/foreman) tool(only the development machine needs it).
4. A clone of the server code in the GOPATH(`git clone git@github.com:larzconwell/moln $GOPATH/src/github.com/larzconwell/moln`).

#### Running a Development Server
1. Run the dev script `./dev`.

This script will build the server, create any directories needed, and start the servers with the
Foreman tool(including the Redis server).

#### Deploying to a Production Server
1. Configure port 80 to reroute to 3000.
2. Create directories `/mnt/www/moln`, `/data`, `/var/log/moln`, and `/var/log/redis`.
3. Create files `/etc/init/redis-server.conf` and `/etc/init/moln.conf`.
4. Make sure the directories and files created above can be written to by the user.
5. Make sure the `redis-server` and `moln` upstart process can be started and stopped by the user.
6. Run the deploy script `./deploy <user@host...>`.

This script will do the following for each given server:
- build the binary for the servers arch(on the development machine)
- copy the binary and the config directory to appropriate place
- copy the upstart scripts to the appropriate place
- restart redis-server, and moln upstart processes

#### Notes:
- The servers being deployed to are assumed to be Linux.
- When running in production mode on Windows, Redis connects to a UNIX Socket; so you may encounter errors.

#### TLS
So the deployment process above is for standard HTTP, if you want to secure your servers traffic,
add the following to your production configuration(`config/production.json`).
```
{
  "TLS": {
    "Key": "/path/to/server.key",
    "Cert": "/path/to/server.crt"
  }
}
```

Then instead of redirecting port 80's traffic to port 3000, redirect port 443.

### Developers
If you're interested in how the API works, or interested in building or contributing to a Moln
client you should read the [API.md](https://raw.github.com/larzconwell/moln/master/API.md) file
included.

### License
MIT licensed, see [here](https://raw.github.com/larzconwell/moln/master/README.md)

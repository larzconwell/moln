Moln
---

Moln is a simple sync API server with features similar to OwnCloud, including multiple user accounts, file sync, and much more.

### Features
- [] Multiple user accounts
- [] Account access from devices
- [] Account access logs
- [] File sync
- [] Calendar sync
- [] Address book sync
- [] Task manager

### Running a server
1. Install Redis stable.
2. Install a [Foreman](https://github.com/ddollar/forego) tool.
3. Clone the server code `git clone git@github.com:larzconwell/moln.git` and `cd` to it.
4. Run the setup script `[sudo] ./setup`, it'll create the appropriate directories(`sudo` is for production).
5. Start the Foreman process(e.g. `forego start`) with the appropriate Procfile(`Procfileprod` for production).

#### Notes:
- When you clone for it to build you'll have to clone into the project directory in GOPATH(e.g. `$GOPATH/src/github.com/larzconwell/moln`).
- When running in production mode on Windows, Redis connects to a UNIX Socket; so you may encounter errors.

### License
MIT licensed, see [here](https://raw.github.com/larzconwell/moln/master/README.md)

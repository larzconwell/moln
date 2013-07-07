Moln
---

Moln is a simple sync API server with features similar to OwnCloud including multiple user accounts, file sync, and much more.

Currently no clients are available though in the future I plain to implement some clients for popular platforms.

### Features
- [] Multiple user accounts
- [] Device access logs
- [] File sync
- [] Calendar sync
- [] Address book sync
- [] Task manager

### Running a server
1. Install Redis 2.6.14
2. Install a [Foreman](https://github.com/ddollar/foreman) tool
3. Get the server code with `go get github.com/larzconwell/moln` or just clone it `git clone git@github.com:larzconwell/moln.git`
4. Run the setup script `sudo ./setup`, it'll create the appropriate directories for development and production
5. Start the forman process(e.g. `foreman start -e config/development.env`) with the correct environment file

Note: When running in production mode on Windows, Redis connects to a UNIX Socket so you may encounter errors

### Redis Data
This is purely here for reference when messing with the server directly

- `users:<user>`
  - `name <user> password <password>`
  - Hash of users data
- `users:<user>:devices`
  - `<device>, ...`
  - Set of users device names
- `users:<user>:devices:<device>`
  - `name <device> token <token> user <user>`
  - Hash of a users device data
- `tokens:<token>`
  - `device <device> user <user>`
  - Hash of a tokens data

### API Documentation

### License
MIT, view the included [LICENSE](https://raw.github.com/larzconwell/moln/master/LICENSE) file for details.

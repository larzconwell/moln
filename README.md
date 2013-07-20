Moln
---

Moln is a simple sync API server with features similar to OwnCloud including multiple user accounts, file sync, and much more.

Currently no clients are available though in the future I plain to implement some clients for popular platforms.

### Features
- [DONE] Multiple user accounts
- [DONE] Account access from devices
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


### License
MIT, view the included [LICENSE](https://raw.github.com/larzconwell/moln/master/LICENSE) file for details.

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
#### Notes

##### Response bodies
If validations fail for POST/PUT requests a 400 response is sent with the following body.
```
{errors: ['']}
```
_Note: This does not mean all 400 responses are validation failures._

If the status code is not 2xx, there were no validation errors, and it's not a redirect code the following is the response body.
```
{error: ''}
```

Reponses with 2xx status codes can have one of the following responses. The route details below will inform you of the response body sent.

- USER: `{name: '', devices: [DEVICE]}`
- DEVICE: `{name: '', token: ''}`

##### Response Content Types
Response content types can be requested when you send a request. Just include a `Content-Type` header with the mime type for the response you want. 

If the mime type is not supported a 415 is returned including the `Accept` header listing the supported mime types.

If no `Content-Type` is sent `application/json` is the default for responses.

POST/PUT requests always send a `application/json` response, that's because the request `Content-Type` for them will be a form type(e.g. `application/x-www-form-urlencoded`) typically.

##### Authentication and Authorization
Authenticating with the server can be done in one of three ways

- A `token` field in the URL query(i.e. `?token=<token>`).
- An `Authorization` header in the following format(i.e. `Authorization: Token <token>`).
- An `Authorization` header including the format `user:password` encoded as base64(i.e. `Authorization: Basic <base64>`).

#### Routes
##### POST /users
Creates a user if available. An initial device is created if `devicename` is given so you have a token to start with

- Data: `name`, `password`, `devicename`
- Body: `{user: USER}`

##### GET /users/{name}
Get a user, if authenticated and the user is the retrieved user the devices token fields are given, otherwise empty

- Authentication: optional
- Body: `{user: USER}`

##### PUT /users/{name}
Update a user, authenticated user must be the user being updated

- Data: `password`
- Authentication: required
- Body: `{user: USER}`

##### DELETE /users/{name}
Delete a user, authenticated user must be the user being deleted

- Authentication: required
- Body: `{user: USER}`

##### GET /users/{user}/devices
Get the users devices, if authenticated and the user is the retrieved user the devices token fields are given, otherwise empty

- Authentication: optional
- Body: `{devices: [DEVICE]}`

##### POST /users/{user}/devices
Create a new device for a user, authenticated user must be the user the device belongs to

- Data: `name`
- Authentication: required
- Body: `{device: DEVICE}`

##### GET /users/{user}/devices/{name}
Get a users device, if authenticated and the user is the retrieved user the devices token field is given, otherwise empty

- Authentication: optional
- Body: `{device: DEVICE}`

##### DELETE /users/{user}/devices/{name}
Delete a users device, authenticated user must be the user the device belongs to

- Authentication: required
- Body: `{device: DEVICE}`

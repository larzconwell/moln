### API
The following response bodies are JSON formatted, but actual responses may be in other formats.

### Authentication and Authorization
Authenticating with Moln can be done in two ways.

1. An `Authorization` header in the following format `Authorization: Token <token>`.
2. An `Authorization` header including the format `user:password` encoded as base64(i.e. `Authorization: Basic <base64>`).

If authentication fails or no authentication is provided where required a `401` is returned.

### Responses
#### Response Formats
You can choose the responses `Content-Type` by either giving an `Accept` header listing the response
formats you accept, or you can give an extension to the path and Moln will respond in the mime
type matched with the extension.

If the extension doesn't match any mime type, or none of the `Accept` header values are supported
the response will be `406`.

Currently, only `application/json` is supported for responses, and if no extension or `Accept`
header are given then it is used.

#### Response Bodies
For POST/PUT requests, validations occur to ensure the data you send can be set correctly.
If any validations fail then a `400` is returned with the following body.
```
{"errors": [""]}
```
_Note: This does not mean all `400` responses are validation failures._

If the status code is not 2xx, there are no validation errors, and it's not a redirect code then
the following error body is returned.
```
{"error": ""}
```

### Routes
In the response sections below for each route, you will see invalid JSON in the format `<NAME>`,
these are snippets and the following snippets defined below should be read in place of the name.
- `USER`: `{"name": ""}`
- `DEVICE`: `{"name": "", "token": ""}`
- `ACTIVITY`: `{"time": "", "message": ""}`

#### Users
##### POST /user
Create a user if available. If the `device` item is given, an initial device is also created;
this is so you don't have to authenticate with the users password.

- Data: `name`, `password`, `device`
- Response:
  - `<USER>` If no `device` item is given.
  - `{"user": <USER>, "device": <DEVICE>}` If a `device` item is given.

##### GET /user
Get the authenticated user.

- Authentication: required
- Response: `{"user": <USER>, "devices": [<DEVICE>], "activities": [<ACTIVITY>]}`

##### PUT /user
Update the authenticated users data.

- Data: `password`
- Authentication: required
- Response: `<USER>`

##### DELETE /user
Delete the authenticated user.

- Authentication: required
- Response: `<USER>`

#### Devices
##### POST /devices
Create a device for the authenticated user if avaiable.

- Data: `name`
- Authentication: required
- Response: `<DEVICE>`

##### GET /devices/{name}
Get a device from the authenticated user.

- Authentication: required
- Response: `<DEVICE>`

##### DELETE /devices/{name}
Delete a device from the authenticated user.

- Authentication: required
- Response: `<DEVICE>`

### Redis
The following list is a reference to the backend Redis keys
- `users:<user>`
  - `name <user> password <password>`
  - A hash of user data
- `users:<user>:devices`
  - `<device>, ...`
  - Set of users device names
- `users:<user>:devices:<device>`
  - `name <device> token <token>`
  - Hash of device data
- `users:<user>:activities`
  - `<activity>, ...`
  - List of activity times
- `users:<user>:activities:<activity>`
  - `time <activity> message <message>`
  - Hash of activity data
- `tokens:<token>`
  - `device <device> user <user>`
  - Hash of token data

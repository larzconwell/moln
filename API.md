### API
The following response bodies are JSON formatted, but actual responses may be in other formats.

### Authentication and Authorization

Authenticating with Moln can be done in three ways
1. A `token` field in the URL query(i.e. `?token=<token>`).
2. An `Authorization` header in the following format `Authorization: Token <token>`.
3. An `Authorization` header including the format `user:password` encoded as base64(i.e. `Authorization: Basic <base64>`).

If authentication is required and the authenticated user is not authorized a `403` is returned.
If authentication fails or no authentication is provided where required a `401` is returned.

### Responses

#### Response Formats
You can choose the responses `Content-Type` by either giving an `Accept` header listing the response
formats you accept, or you can give an extension to the path and Moln will respond in the mime
type matched with the extension.

Unsupported formats you request will respond with a `406`.

Currently only `application/json` is supported for responses, and if no extension or `Accept`
header then it is used.

#### Response Bodies
For POST/PUT requests, validations occur to ensure the data your sending can be set correctly.
If any validations fail then a `400` is returned with the following body.
```
{"errors": [""]}
```
_Note: This does not mean all `400` responses are validation failures._

If the status code is not 2xx, there were no validation errors, and it's not a redirect code then
the following error body is returned.
```
{"error": ""}
```

2xx Responses may have one of the following bodies, the route details below will inform you
which is given.
# TODO

### Routes

### Oct 18, 2013
- Use the default Redis maxmemory policy, volatile-lru instead of volatile-ttl
- Create user activity if failed password match occurs for Basic auth

### Oct 17, 2013
- Increment task ID using Redis incr command

### Oct 16, 2013
- Basic and Token authentication comply with rfc dealing with realm param.

### Oct 13, 2013
- Add categories to tasks

### Oct 12, 2013
- Requests now use a Redis connection pool
- Expose server config as global
- Add time to uuid when generating tokens

### Oct 09, 2013
- Add device set retrieval handler
- Add activities retrieval handler
- Add handlers to retrieve tasks
- Complete task removal handler
- When removing user remove tasks
- Complete task creation handler
- Complete update task handler

### Oct 06, 2013
- Complete remove user handler
- Implement user and device removal
- Complete all device handlers
- Complete user activity logging

### Oct 05, 2013
- Complete user create handler
- Complete update user handler
- Complete get user handler
- Token struct to manage tokens
- Remove option to authenticate with query param
- Implement Basic and Token authentications
- Implement user and device retrieval

### Oct 04, 2013
- Replace Validate with Validations and add HandlerValidations to manage HTTP
- Add db code to validate and save user data

### Oct 03, 2013
- Add code to connect to the Redis db

### Sep 24, 2013
- Create routing mechanisms
- Add validation errors
- Add validation function
- Add user create handler that validates data

### Sep 21, 2013
- Create upstart scripts for moln and redis
- db directory now data
- log files have .log
- add dev script replacing setup
- add deploy script
- move not found handler calls to main server handler

### Sep 19, 2013
- Refactor configurations
- Use Procfiles to run servers
- Setup script should only create log dir if superuser
- Get server config depending on environment
- Add router and basic http server

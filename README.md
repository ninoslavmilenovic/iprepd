# iprepd

iprepd is a centralized reputation daemon that can be used to store reputation information
for various objects such as IP addresses and retrieve reputation scores for the objects.

The project initially focused on managing reputation information for only IP addresses, but
has since been expanded to allow reputation tracking for other types of values such as account
names or email addresses.

The daemon provides an HTTP API for requests, and uses a Redis server as the backend storage
mechanism. Multiple instances of the daemon can be deployed using the same Redis backend.

__Note__: Support for legacy endpoints, such as `GET /127.0.0.1` has been discontinued. Typed
endpoints must now be used. For more information see the API documentation below. Some minor
modification to client code is required to convert from the old endpoints.

## Configuration

Configuration is done through the configuration file, by default `./iprepd.yaml`. The location
can be overridden with the `-c` flag.

See [iprepd.yaml.sample](./iprepd.yaml.sample) for an example configuration.

## Building the Docker image

```bash
./write_version_json.sh
docker build -t iprepd:latest .
```

Docker images are also [published](https://hub.docker.com/r/mozilla/iprepd/).

```bash
docker pull mozilla/iprepd:latest
docker run -ti --rm -v `pwd`/iprepd.yaml:/app/iprepd.yaml mozilla/iprepd:latest
```

## API

### Authentication

iprepd supports two forms of authentication. Clients can authenticate to the service using either
standard API keys, or by using [Hawk authentication](https://github.com/hapijs/hawk).

Standard API key authentication can be configured in the configuration file in the `apikey` (for
read/write) and `ROapikey` (for a read-only user) sections under the `auth` section in the
configuration file. To use API key authentication, clients should send an `Authorization` header
that is of format `APIKey <apikey>`.

Similarly, Hawk authentication is configured under the `hawk` and `ROhawk` sections in the
configuration file. To use Hawk authentication clients need to include the hawk authentication
header in the `Authorization` header when making a request.

### Endpoints

#### GET /type/ip/10.0.0.1

Request the reputation for an object of a given type. Responds with 200 and a JSON
document describing the reputation if found. Responds with a 404 if the object is
unknown to iprepd, or is in the exceptions list.

The current supported object types are `ip` for an IP address and `email` for an
email address.

The response body may include a `decayafter` element if the reputation for the address was changed
with a recovery suppression applied. If the timestamp is present, it indicates the time after which
the reputation for the address will begin to recover.

##### Response body

```json
{
	"object": "10.0.0.1",
        "type": "ip",
	"reputation": 75,
	"reviewed": false,
	"lastupdated": "2018-04-23T18:25:43.511Z"
}
```

#### DELETE /type/ip/10.0.0.1

Deletes the reputation entry for the requested object of the specified type.

#### PUT /type/ip/10.0.0.1

Sets a reputation score for the specified object of the specified type. A reputation JSON
document must be provided with the request body. The `reputation` field must be provided
in the document. The reviewed field can be included and set to true to toggle the reviewed
field for a given reputation entry.

Note that if the reputation decays back to 100, if the reviewed field is set on the entry it will
toggle back to false.

The reputation will begin to decay back to 100 immediately for the address based on the decay
settings in the configuration file. If it is desired that the reputation should not decay for a
period of time, the `decayafter` field can be set with a timestamp to indicate when the reputation
decay logic should begin to be applied for the entry.

##### Request body

```json
{
	"object": "10.0.0.1",
        "type": "ip",
	"reputation": 75
}
```

#### PUT /violations/type/ip/10.0.0.1

Applies a violation penalty to the specified object of the specified type.

If an unknown violation penalty is submitted, this endpoint will still return 200, but the
error will be logged.

If desired, `suppress_recovery` can be included in the request body and set to an integer which
indicates the number of seconds that must elapse before the reputation for this entry will begin
to decay back to 100. If this setting is not included, the reputation will begin to decay
immediately. If the violation is being applied to an existing entry, the `suppress_recovery` field
will only be applied if the existing entry has no current recovery suppression, or the specified
recovery suppression time frame would result in a time in the future beyond which the entry
currently has. If `suppress_recovery` is included it must be less than `1209600` (14 days).

##### Request body

```json
{
	"object": "10.0.0.1",
        "type": "ip",
	"violation": "violation1"
}
```

#### PUT /violations/type/ip

Applies a violation penalty to a multiple objects of a given type.

If an unknown violation penalty is submitted, this endpoint will still return 200, but the
error will be logged.

##### Request body

```json
[
	{"object": "10.0.0.1", "type": "ip", "violation": "violation1"},
	{"object": "10.0.0.2", "type": "ip", "violation": "violation1"},
	{"object": "10.0.0.3", "type": "ip", "violation": "violation2"}
]
```

#### GET /violations

Returns violations configured in iprepd in a JSON document.

##### Response body

```json
[
	{"name": "violation1", "penalty": 5, "decreaselimit": 50},
	{"name": "violation2", "penalty": 25, "decreaselimit": 0},
]
```

#### GET /dump

Returns all reputation entries.

**Note: This makes use of the [KEYS](https://redis.io/commands/keys) redis command, which is known to be very slow. Use with care.**

##### Response body

```json
[
  {"ip": "10.0.0.1", "reputation": 75, "reviewed": false, "lastupdated": "2018-04-23T18:25:43.511Z"},
  {"ip": "10.0.0.2", "reputation": 50, "reviewed": false, "lastupdated": "2018-04-23T18:31:27.457Z"},
  {"ip": "10.0.20.2", "reputation": 25, "reviewed": false, "lastupdated": "2018-04-23T17:22:42.230Z"},
]
```


#### GET /\_\_heartbeat\_\_

Service heartbeat endpoint.

#### GET /\_\_lbheartbeat\_\_

Service heartbeat endpoint.

#### GET /\_\_version\_\_

Return version data.

## Acknowledgements

The API design and overall concept for this project are based on work done in
[Tigerblood](https://github.com/mozilla-services/tigerblood).

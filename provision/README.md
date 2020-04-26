# Provision service

Provision service provides an HTTP API to interact with Mainflux. 
Provision service is used to setup initial applications configuration i.e. things, channels, connections and certificates that will be required for the specific use case. 
For example lets say you are using a Mainflux and connecting gateways to it. Gateways need an easy way to configure itself for communication with Mainflux (receiving and sending controls and data) and for that you will use bootstrap service but before gateway connects to the bootstrap service configuration needs to be created (things, channels, connections, bootstrap configuration and certificates if mtls is being used). Provision service should provide an easy way of provisioning your gateways. On a gateway there can be many services running and you may require more than one thing and channel to establish all the required means of representing a gateway in your application. Let's say that you are using an [Agent](https://github.com/mainflux/agent) and [Export](https://github.com/mainflux/export) service and you use mtls you will need to provision two things and certificates for those things for access to Mainflux, or you may want to create any number of things and channels for your application, this kind of setup we can call provision layout.
Provision service provides this feature on a “/mapping” endpoint. Provision layout is configured in [config.toml](configs/config.toml)

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                            | Description                                       | Default                          |
| ----------------------------------- | ------------------------------------------------- | -------------------------------- |
| MF_PROVISION_USER                   | User (email) for accessing Mainflux               | user@example.com                 |
| MF_PROVISION_PASS                   | Mainflux password                                 | user123                          |
| MF_PROVISION_API_KEY                | Mainflux authentication token                     | ""                               |
| MF_PROVISION_CONFIG_FILE            | Provision config file                             | "config.toml"                    |
| MF_PROVISION_HTTP_PORT              | Provision service listening port                  | 8091                             |
| MF_PROVISION_ENV_CLIENTS_TLS        | Mainflux SDK TLS verification                     | false                            |
| MF_PROVISION_CA_CERTS               | Mainflux gRPC secure certs                        | ""                               |
| MF_PROVISION_SERVER_CERT            | Mainflux gRPC secure server cert                  | ""                               |
| MF_PROVISION_SERVER_KEY             | Mainflux gRPC secure server key                   | ""                               |
| MF_PROVISION_SERVER_KEY             | Mainflux gRPC secure server key                   | ""                               |
| MF_PROVISION_MQTT_URL               | Mainflux MQTT adapter URL                         | "http://localhost:1883"          |
| MF_PROVISION_USERS_LOCATION         | Users service URL                                 | "http://locahost"                |
| MF_PROVISION_THINGS_LOCATION        | Things service URL                                | "http://localhost"               |
| MF_PROVISION_LOG_LEVEL              | Service log level                                 | "http://localhost"               |
| MF_PROVISION_HTTP_PORT              | Service listening port                            | "8091"                           |
| MF_PROVISION_USER                   | Mainflux user username                            | "test@example.com"               |
| MF_PROVISION_PASS                   | Mainflux user password                            | "password"                       |
| MF_PROVISION_BS_SVC_URL             | Mainflux Bootstrap service URL                    | http://localhost/things/configs" |
| MF_PROVISION_BS_SVC_WHITELIST_URL   | Mainflux Bootstrap service whitelist URL          | "http://localhost/things/state"  |
| MF_PROVISION_CERTS_SVC_URL          | Certificats service URL                           | "http://localhost/certs"         |
| MF_PROVISION_X509_PROVISIONING      | Should X509 client cert be provisioned            | "false"                          |
| MF_PROVISION_BS_CONFIG_PROVISIONING | Should thing config be saved in Bootstrap service | "true"                           |
| MF_PROVISION_BS_AUTO_WHITEIST       | Should thing be auto whitelisted                  | "true"                           |
| MF_PROVISION_BS_CONTENT             | Bootstrap service content                         | "{}"                             |

By default, call to `/mapping` endpoint will create one thing and two channels (`control` and `data`) and connect it. If there is a requirement for different provision layout we can use [config](docker/configs/config.toml) file in addition to environment variables. For the purposes of running provision as an add-on in docker composition environment variables seems more suitable. Environoment variables are set in [.env](.env).  
Configuration can be specified in [config.toml](configs/config.toml). Config file can specify all the settings that environment variables can configure and in addition
`/mapping` endpoint provision layout can be configured.

In `config.toml` we can enlist array of things and channels that we want to create and make connections between them which we call provision layout.
Metadata can be whatever suits your needs except that at least one thing needs to have `externalID` (which is populated with value from [request](#example)).
For channels metadata `type` is reserved for `control` and `data` which we use with [Agent](https://github.com/mainflux/agent).

Example below
```
[[things]]
  name = "thing"

  [things.metadata]
    externalID = "xxxxxx"


[[channels]]
  name = "control-channel"

  [channels.metadata]
    type = "control"

[[channels]]
  name = "data-channel"

  [channels.metadata]
    type = "data"

[[channels]]
  name = "export-channel"

  [channels.metadata]
    type = "data"
```

## Authentication
In order to create necessary entities provision service needs to authenticate against Mainflux api. To provide authentication credentials to the provision service you can pass it in an environment variable or in a config file as Mainflux user and password or as api token (that can be issued on users/keys endpoint). Additionally users or api token can be passed in Authorization header, this authentication takes precedence over others.
* MFUser, MFPass  in config or environment
* MFAPiKEy in config or environment
* Authorization: Token|ApiKey


## Running

Provision service can be run as a standalone or in docker composition as addon to the core docker composition.
```
docker-compose -f docker/addons/provision/docker-compose.yml up
```
or
```
MF_PROVISION_BS_SVC_URL=http://localhost:8202/things MF_PROVISION_THINGS_LOCATION=http://localhost:8182 MF_PROVISION_USERS_LOCATION=http://localhost:8180 MF_PROVISION_CONFIG_FILE=docker/addons/provision/configs/config.toml build/mainflux-provision
```

For the case that credentials or api token is passed in configuration
```
curl -s -S  -X POST  http://localhost:8888/mapping  -H 'Content-Type: application/json' -d '{ "external_id" : "33:52:77:99:43", "external_key":"223334fw2" }'
```

In the case that provision service is not deployed with credential or you want to use user other than default one being set in environment ( or config file)
```
curl -s -S  -X POST  http://localhost:8091/mapping -H "Authorization: $TOK" -H 'Content-Type: application/json'   -d '{ "external_id" : "02:42:fE:65:D3:23", "external_key":"223334fw2" }'
```

Response contains created things and channels and certificates if any.
```
{
  "things": [
    {
      "ID": "c22b0c0f-8c03-40da-a06b-37ed3a72c8d1",
      "Owner": "",
      "Name": "thing",
      "Key": "007cce56-e0eb-40d6-b2b9-ed348a97d1eb",
      "Metadata": {
        "externalID": "33:52:79:C3:43"
      }
    }
  ],
  "channels": [
    {
      "ID": "064c680e-181b-4b58-975e-6983313a5170",
      "Owner": "",
      "Name": "control-channel",
      "Metadata": {
        "type": "control"
      }
    },
    {
      "ID": "579da92d-6078-4801-a18a-dd1cfa2aa44f",
      "Owner": "",
      "Name": "data-channel",
      "Metadata": {
        "type": "data"
      }
    }
  ],
  "whitelisted": {
    "c22b0c0f-8c03-40da-a06b-37ed3a72c8d1": true
  }
}
```


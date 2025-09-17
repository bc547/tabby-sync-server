> [!IMPORTANT]
> The included web interface is broken due to changes to the CDN that hosted the javascript libraries. Fixing this properly would require implementing the frontend website again. I currently don't have the time to implement this again :-(
>
> Even tough the frontend website is broken, the admin API calls mentioned in the README (https://github.com/bc547/tabby-sync-server?tab=readme-ov-file#admin-api-overview) still work. 
> 
> This repo is now archived to make clear this code is not really maintained anymore at this moment.

# tabby-sync-server

Small and simple configuration synchronization server for Tabby with support for multiple users (synctokens) and Tabby
configs. It includes a built-in admin website to manage the synctokens.

<kbd>![](./docs/media/adminpage.png)</kbd>

### Some features

* Runs as a single container (or executable)
* Built-in admin portal website
  * Requires an admin token to view/create/delete syntokens
  * Can easily be completely disabled for extra security
* Built-in database that requires no extra setup or configuration (bbolt db)
* Very simple configuration
  * Generate admin token for accessing the admin API (environment variable)
  * Generate .Tabby synctokens via built-in admin portal (a few mouse clicks)
  * Configure Tabby (fill in two values and press some buttons)
* Securing the API connection can be done by any frontend (web)proxy of your own choice
* Very small size and footprint: docker image is less than 4 MBytes

# Quick start

Assume tabby-sync-server runs on the same host as Tabby. So it is directly reachable via the address 127.0.0.1

### Run tabby-sync-server

* Generate a UUIDv4 identifier (E.g. on https://www.uuidgenerator.net/version4).
  E.g. ```f87f61f8-b173-485f-bc04-9655c74afc0d```
  This identifier is used as authorization token for the admin API and should be kept secret.


* Start the container
  ```bash
  docker run --name tabby-sync-server -v ./tabby-sync:/data -e "ADMIN_KEY=f87f61f8-b173-485f-bc04-9655c74afc0d" -e "LOG_LEVEL=DEBUG" -p 8080:8080 proplab/tabby-sync-server:latest 
  ```

  **Remark: Running like this in a production setup is not recommended since the connection is not encrypted. See "Installation->Docker compose"
below for a more secure way with a frontend proxy.**

### Create new synctoken

* Surf to http://127.0.0.1/admin to access the admin portal. Click on 'Admin' and enter the configured admin token.


* Click on "Create Sync Token" and copy the generated information.
  <kbd>![](docs/media/createpopup.png)</kbd>

### Configure Tabby

* Open settings and goto 'Config Sync'


* Fill in the required parameters
  * ```Sync host``` The url to the tabby-sync-server. If running on the same host eg. http://127.0.0.1:8080
  * ```Secret Sync Token``` The new synctoken generated in the step above


* Click on ```Upload as new config``` and choose a name for your configuration. Your configuration is now uploaded and visible as a config


* Turn on ```Sync automatically```

# Configuration

This service is designed to run in a containerized environment. All settings are provided via environment variables and
no config file is needed.

### Sync host setting in Tabby
It is possible to add extra information to the Sync host URL field in Tabby.


This information will be visible in the logs of tabby-sync-service and can be used to distinguish individual clients (e.g. for debugging).

**Example:**

Suppose tabby-sync-service listens on https://tabby-sync.example.net.

On my laptop I can enter https://tabby-sync.example.net/mylaptop as sync host, and on my desktop https://tabby-sync.example.net/mydesktop

In the tabby-sync-server logs, there will be an extra field *clientinfo* that contains the specific tabby instance making the request or change (e.g. clientinfo="mydesktop").

**Note: The max length is 32 characters and special characters are stripped or escaped.**

### Environment settings overview

| Variable     | Default              | Description                                                                                                                                 |
|--------------|----------------------|---------------------------------------------------------------------------------------------------------------------------------------------|
| LOG_LEVEL    | INFO                 | Verbosity of the logs. Possible values are: DEBUG,INFO,WARNING,ERROR. DEBUG level will show all received http requests                      |
| ADMIN_KEY    | -                    | If set, the admin API will be enabled and allow to add/remove/view synctokens. If not set, the admin portal and API are completely disabled |
| DB_FILE      | /data/configstore.db | Name and location of the database file (bbolt db) inside the container. Normally there is no reason to change this                          |
| HTTP_ADDRESS | :8080                | Address on which the server will listen                                                                                                     |

**For new installations, it is mandatory to at least configure ADMIN_KEY temporarily and add at least 1 synctoken**

### Admin API overview

Authentication is performed by means of the Authorization HTTP header with the ADMIN_KEY as bearer token

| Path                      | Method | Description                                                              |
|---------------------------|--------|--------------------------------------------------------------------------|
| /api/admin/synctokens     | GET    | Get a list of all synctokens                                             |
| /api/admin/synctokens     | POST   | Create a new synctoken. The id is generated and returned in the response |
| /api/admin/synctokens/:id | DELETE | Delete the specified synctoken and all the associated configs            |

**Remark: The admin API is only available if ADMIN_KEY is set**

### Examples

* Add new synctoken
  ```shell
  curl https://tabby-sync.example.net/api/admin/synctokens -XPOST -H "Authorization: bearer f87f61f8-b173-485f-bc04-9655c74afc0d"
  ```
  returns
  ```json
  {
    "msg": "synctoken created",
    "synctoken": "62a28a5e-5bb3-41f9-a2fb-12682e5f12cd"
  }
  ```

* List all synctokens
  ```shell
  curl https://tabby-sync.example.net/api/admin/synctokens -H "Authorization: bearer f87f61f8-b173-485f-bc04-9655c74afc0d"
  ```
  returns
  ```json
  {
    "msg": "List of all synctokens",
    "synctokens": [
      "62a28a5e-5bb3-41f9-a2fb-12682e5f12cd",
      "6be5c35e-fa60-4460-bebc-01fba0e47c95"
    ]
  }
  ```

* Delete synctoken
  ```shell
  curl https://tabby-sync.example.net/api/admin/synctokens/62a28a5e-5bb3-41f9-a2fb-12682e5f12cd -XDELETE -H "Authorization: bearer f87f61f8-b173-485f-bc04-9655c74afc0d"
  ```
  returns
  ```json
  {
    "msg": "synctoken deleted",
    "synctoken": "62a28a5e-5bb3-41f9-a2fb-12682e5f12cd"
  }
  ```

# Installation

### Docker

```bash
docker run --name tabby-sync-server -v ./tabby-sync:/data -e "ADMIN_KEY=f87f61f8-b173-485f-bc04-9655c74afc0d" -e "LOG_LEVEL=DEBUG" proplab/tabby-sync-server:latest 
```

**Remark: Not recommended**

### Docker-compose (with traefik as frontend proxy)

```yaml
  tabby-sync-server:
    hostname: "tabby-sync-server"
    container_name: "tabby-sync-server"
    image: "proplab/tabby-sync-server:latest"
    environment:
      - "ADMIN_KEY=f87f61f8-b173-485f-bc04-9655c74afc0d"
    volumes:
      - "./tabby-sync-server:/data"
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.tabby_sync_server.rule=Host(`tabby-sync.example.net`)"
      - "traefik.http.routers.tabby_sync_server.entrypoints=websecure"
    restart: "unless-stopped"
    logging:
      driver: journald
      options:
        tag: "docker-{{.Name}}"
```

**Remarks:**

- Frontend proxy configuration (traefik) is not shown since it depends on your own setup and preferences. But examples can be found online (e.g. https://doc.traefik.io/traefik/user-guides/docker-compose/acme-tls/)


  In my case:
  - Traefik terminates the https connection and provides free TLS certificates (ZeroSSL) to properly setup a secure connection
  - The service is directly reachable via https://tabby-sync.example.net
  - This container will log everything directly to journald (and not to the normal docker log target). This is not mandatory and
  can be omitted.

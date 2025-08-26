# Sarracenia

[![AGPLv3 License](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/CTAG07/Sarracenia)](https://goreportcard.com/report/github.com/CTAG07/Sarracenia)
[![Go Version](https://img.shields.io/github/go-mod/go-version/CTAG07/Sarracenia)](https://golang.org)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/CTAG07/Sarracenia)](https://github.com/CTAG07/Sarracenia/releases/latest)
[![Docker Image](https://img.shields.io/badge/ghcr.io-ctag07/sarracenia-blue?logo=docker)](https://github.com/CTAG07/Sarracenia/pkgs/container/sarracenia)
[![Repo size](https://img.shields.io/github/repo-size/CTAG07/Sarracenia)](https://github.com/CTAG07/Sarracenia)

[![Sarracenia Test / Build / Release](https://github.com/CTAG07/Sarracenia/actions/workflows/go.yml/badge.svg)](https://github.com/CTAG07/Sarracenia/actions/workflows/go.yml)
[![CodeQL Advanced](https://github.com/CTAG07/Sarracenia/actions/workflows/codeql.yml/badge.svg)](https://github.com/CTAG07/Sarracenia/actions/workflows/codeql.yml)

A Go-based anti-scraper tarpit inspired by Nepenthes.

The goal of Sarracenia is to serve as a defensive wall against web scrapers by generating endless, plausible-looking web
pages, to prevent them from accessing protected content by keeping them in a loop of fake ones.

---

## Reusable Components

The core components of Sarracenia are standalone, reusable libraries in the `/pkg` directory.

### `pkg/markov` - Go Markov Library

A high-performance library for training and using Markov chain models. It is feature-complete and includes a streaming
API, database persistence, and advanced generation features.

**[➡️ View the full README for the Markov Library](./pkg/markov/README.md)**

### `pkg/templating` - Dynamic HTML Template Engine

The rendering engine for Sarracenia. It uses text from the Markov library and/or a wordlist to embed content within
complex, randomized HTML structures.

**[➡️ View the full README for the Templating Library](./pkg/templating/README.md)**

---

## Installation and Usage

There are four primary ways to install and run Sarracenia. After installation, proceed to the **"Initial Setup"**
section, which is common to all methods.

### 1. From a Pre-compiled Release (Recommended for most users)

This is the simplest method. It uses the pre-compiled binaries provided with each new version on GitHub.

**Steps:**

1. **Download the Latest Release:**
    * Go to the [**Sarracenia Releases Page**](https://github.com/CTAG07/Sarracenia/releases/latest).
    * Find the asset that matches your operating system and architecture (e.g., `sarracenia-linux-amd64`,
      `sarracenia-windows-amd64.exe`).
    * Download the binary.

2. **Prepare Initial Data:**
    * From the same releases page, download the `Source code (zip)` or `Source code (tar.gz)` file.
    * Extract the archive. You will need the contents of the `example` directory.
    * Create your application directory and arrange the files as follows:

      ```
      /your/sarracenia/folder/
      ├── sarracenia              # The binary you downloaded
      ├── config.json             # Copied from the extracted 'example' folder
      └── data/
          ├── dashboard/
          │   ├── static/
          │   └── templates/
          ├── templates/          # Copied from 'example/data'
          │   └── ...
          └── wordlist.txt        # Copied from 'example/data'
      ```

3. **Run the Application:**
    * **On Linux or macOS:**
      ```sh
      # Make the binary executable
      chmod +x ./sarracenia

      # Run the application
      ./sarracenia
      ```
    * **On Windows:**
        * Simply double-click the `sarracenia.exe` file or run it from the command prompt:
      ```sh
      .\sarracenia.exe
      ```

### 2. From Source (All Platforms)

This method is suitable for developers or for platforms where Docker or systemd are not available.

**Prerequisites:**

* Go 1.24.5 or later.
* Git for cloning the repository.

**Steps:**

1. **Clone the Repository:**
   ```sh
   git clone https://github.com/CTAG07/Sarracenia.git
   cd Sarracenia
   ```

2. **Build and Run:**
   ```sh
   # Build the application
   go build -o sarracenia ./cmd/main

   # Run the application
   ./sarracenia
   ```
   On the first run, Sarracenia will automatically create a `config.json` and a `data` directory with the necessary
   files by copying them from the `example` directory.

### 3. Docker (Recommended for Deployment)

Pre-built Docker images are automatically published to the GitHub Container Registry (`ghcr.io`) with each new release.

1. **Create a `docker-compose.yml` file:**
   ```yaml
   # docker-compose.yml
   services:
     sarracenia:
       image: ghcr.io/ctag07/sarracenia:latest
       container_name: sarracenia
       restart: unless-stopped
       ports:
         - "7277:7277" # Tarpit server
         - "7278:7278" # Dashboard/API
       volumes:
         # Persists your database, config, and templates
         - sarracenia-data:/app/data

   volumes:
     sarracenia-data:
   ```

2. **Start the Service:**
   ```sh
   docker compose up -d
   ```

### 4. Systemd Service (Linux)

For Linux servers, an installation script is provided to set up Sarracenia as a systemd service.

1. **Clone the Repository:**
   ```sh
   git clone https://github.com/CTAG07/Sarracenia.git
   cd Sarracenia
   ```

2. **Run the Installation Script:**
   The script builds the binary, creates a dedicated user, and installs the application to `/opt/sarracenia`.
   ```sh
   sudo bash ./scripts/install_systemd.sh
   ```

3. **Manage the Service:**
    * Check status: `systemctl status sarracenia.service`
    * View logs: `journalctl -u sarracenia.service -f`

## Initial Setup

1. **Access the Dashboard:**
    * By default, the tarpit runs on port `:7277` and the API/Dashboard runs on port `:7278`.
    * Open a web browser and navigate to `http://localhost:7278`.

2. **Create the Master API Key:**
    * The first time you access the dashboard, the API is unsecured.
    * Navigate to the **API Keys** page.
    * Create your first key. This key will automatically be assigned the master (`*`) scope, giving it full permissions.
    * **Copy this key immediately.** It will not be shown again.
    * Once created, all API endpoints (including the dashboard) will be secured.

## Configuration (`config.json`)

Sarracenia is configured using a `config.json` file in the same directory as the executable.

### `server_config`

| Key                     | Description                                                                                | Default                                  |
|-------------------------|--------------------------------------------------------------------------------------------|------------------------------------------|
| `server_addr`           | Address for the tarpit server to listen on.                                                | `:7277`                                  |
| `api_addr`              | Address for the API and dashboard server.                                                  | `:7278`                                  |
| `log_level`             | Logging level (`debug`, `info`, `warn`, `error`).                                          | `info`                                   |
| `data_dir`              | Path to the data directory.                                                                | `./data`                                 |
| `database_path`         | Path to the SQLite database file.                                                          | `./data/sarracenia.db?_journal_mode=WAL` |
| `dashboard_tmpl_path`   | Path to the dashboard GoHTML template files.                                               | `./data/dashboard/templates/`            |
| `dashboard_static_path` | Path to the dashboard static assets (CSS, JS).                                             | `./data/dashboard/static/`               |
| `enabled_templates`     | A list of `.tmpl.html` files to use for the tarpit. If empty, a random template is chosen. | `["page.tmpl.html"]`                     |
| `tarpit_config`         | Settings for response delaying (see below).                                                |                                          |

**`tarpit_config` object:**

| Key                  | Description                                  | Default |
|----------------------|----------------------------------------------|---------|
| `enable_drip_feed`   | If true, sends the response in slow chunks.  | `false` |
| `initial_delay_ms`   | Time to wait before sending the first byte.  | `0`     |
| `drip_feed_delay_ms` | Time to wait between sending chunks.         | `500`   |
| `drip_feed_chunks`   | Number of chunks to split the response into. | `10`    |

### `template_config`

This object configures the templating engine. See the full documentation in [
`pkg/templating/README.md`](./pkg/templating/README.md).

### `threat_config`

This object configures the threat assessment system, allowing you to control how aggressively the tarpit responds based
on client behavior.

| Key                  | Description                                                 | Default |
|----------------------|-------------------------------------------------------------|---------|
| `base_threat`        | The starting threat score for any request.                  | `0`     |
| `ip_hit_factor`      | Value added to score for each hit from an IP.               | `1.0`   |
| `ua_hit_factor`      | Value added to score for each hit from a User Agent.        | `0.5`   |
| `ip_hit_rate_factor` | Multiplier for hits-per-minute from an IP.                  | `10.0`  |
| `ua_hit_rate_factor` | Multiplier for hits-per-minute from a User Agent.           | `5.0`   |
| `max_threat`         | The absolute maximum threat score.                          | `1000`  |
| `fallback_level`     | The threat stage (0-4) to use if no other threshold is met. | `0`     |
| `stages`             | Defines the score thresholds for each threat stage.         |         |

## API Reference

All API endpoints are prefixed with `/api` and require an API key sent in the `sarr-auth` header.

### Authentication (`/api/auth`)

| Method   | Endpoint              | Scope         | Description                                                            |
|----------|-----------------------|---------------|------------------------------------------------------------------------|
| `GET`    | `/api/auth/me`        | *any*         | Checks the validity of the current key and returns its scopes.         |
| `GET`    | `/api/auth/keys`      | `auth:manage` | Lists all API keys (without the raw key).                              |
| `POST`   | `/api/auth/keys`      | `auth:manage` | Creates a new API key. The first key created is always a master key.   |
| `DELETE` | `/api/auth/keys/{id}` | `auth:manage` | Deletes an API key by its ID. The master key (ID 1) cannot be deleted. |

### Markov Models (`/api/markov`)

| Method   | Endpoint                             | Scope          | Description                                                                                            |
|----------|--------------------------------------|----------------|--------------------------------------------------------------------------------------------------------|
| `GET`    | `/api/markov/models`                 | `markov:read`  | Lists all available Markov models and their info.                                                      |
| `POST`   | `/api/markov/models`                 | `markov:write` | Creates a new, empty Markov model.                                                                     |
| `DELETE` | `/api/markov/models/{name}`          | `markov:write` | Deletes a model and all its data.                                                                      |
| `POST`   | `/api/markov/models/{name}/train`    | `markov:write` | Trains a model with a plain text corpus file in the request body.                                      |
| `POST`   | `/api/markov/models/{name}/prune`    | `markov:write` | Prunes a model's chain data based on a minimum frequency.                                              |
| `GET`    | `/api/markov/models/{name}/export`   | `markov:read`  | Exports a model as a JSON file.                                                                        |
| `POST`   | `/api/markov/models/{name}/generate` | `markov:read`  | Generates text from a given markov model with given params                                             |
| `POST`   | `/api/markov/import`                 | `markov:write` | Imports a model from a JSON file in the request body.                                                  |
| `POST`   | `/api/markov/vocabulary/prune`       | `markov:write` | Prunes the global vocabulary of rare tokens across all models.                                         |
| `GET`    | `/api/markov/training/status`        | `markov:read`  | Gives a json response indicating whether training is occurring, and if so, what model is being trained |

### Server Control (`/api/server`)

| Method | Endpoint               | Scope            | Description                                                |
|--------|------------------------|------------------|------------------------------------------------------------|
| `GET`  | `/api/health`          | *none*           | Unauthenticated health check endpoint.                     |
| `GET`  | `/api/server/version`  | `stats:read`     | Returns the application's build version, commit, and date. |
| `GET`  | `/api/server/config`   | `server:config`  | Retrieves the current `config.json`.                       |
| `PUT`  | `/api/server/config`   | `server:config`  | Updates and saves the `config.json`.                       |
| `POST` | `/api/server/restart`  | `server:control` | Initiates a graceful server restart.                       |
| `POST` | `/api/server/shutdown` | `server:control` | Initiates a graceful server shutdown.                      |

### Statistics (`/api/stats`)

| Method   | Endpoint                     | Scope            | Description                                                   |
|----------|------------------------------|------------------|---------------------------------------------------------------|
| `GET`    | `/api/stats/summary`         | `stats:read`     | Gets a high-level summary of total requests, unique IPs, etc. |
| `GET`    | `/api/stats/top_ips`         | `stats:read`     | Lists the top 100 most frequent IP addresses.                 |
| `GET`    | `/api/stats/top_user_agents` | `stats:read`     | Lists the top 100 most frequent User Agents.                  |
| `DELETE` | `/api/stats/all`             | `server:control` | **Deletes all collected statistics.**                         |

### Templates (`/api/templates`)

| Method   | Endpoint                 | Scope             | Description                                                       |
|----------|--------------------------|-------------------|-------------------------------------------------------------------|
| `GET`    | `/api/templates`         | `templates:read`  | Lists the names of all loaded template and partial files.         |
| `GET`    | `/api/templates/{name}`  | `templates:read`  | Gets the raw content of a template file.                          |
| `PUT`    | `/api/templates/{name}`  | `templates:write` | Updates or creates a template file with the request body content. |
| `DELETE` | `/api/templates/{name}`  | `templates:write` | Deletes a template file.                                          |
| `POST`   | `/api/templates/refresh` | `templates:write` | Manually reloads all templates from disk.                         |
| `POST`   | `/api/templates/test`    | `templates:read`  | Tests template syntax from the request body without saving.       |
| `GET`    | `/api/templates/preview` | `templates:read`  | Renders a preview of a saved template with a given threat level.  |

### Whitelist (`/api/whitelist`)

| Method   | Endpoint                   | Scope             | Description                               |
|----------|----------------------------|-------------------|-------------------------------------------|
| `GET`    | `/api/whitelist/ip`        | `whitelist:read`  | Lists all whitelisted IP addresses.       |
| `POST`   | `/api/whitelist/ip`        | `whitelist:write` | Adds an IP address to the whitelist.      |
| `DELETE` | `/api/whitelist/ip`        | `whitelist:write` | Removes an IP address from the whitelist. |
| `GET`    | `/api/whitelist/useragent` | `whitelist:read`  | Lists all whitelisted User Agents.        |
| `POST`   | `/api/whitelist/useragent` | `whitelist:write` | Adds a User Agent to the whitelist.       |
| `DELETE` | `/api/whitelist/useragent` | `whitelist:write` | Removes a User Agent from the whitelist.  |

---

## License

This project is licensed under the AGPLv3.

**Alternative Licensing**

I understand that the AGPL-3.0 may not be suitable for all users. If you would like to use this project in a way not
permitted by the AGPL-3.0 (for instance, in a closed-source application), I am happy to grant you a free, permissive
license (such as the MIT License).

Please contact me at **`82781942+CTAG07@users.noreply.github.com`** to request an alternative license.

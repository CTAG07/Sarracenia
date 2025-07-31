# Sarracenia

[![AGPLv3 License](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/CTAG07/Sarracenia)](https://goreportcard.com/report/github.com/CTAG07/Sarracenia)

A Go-based anti-scraper tarpit inspired by Nepenthes.

The goal of Sarracenia is to serve as a defensive wall against web scrapers by generating endless, plausible-looking web
pages, to prevent them from accessing protected content by keeping them in a loop of fake ones.

## Installation and Usage

There are three primary ways to install and run Sarracenia. After installation, proceed to the **"Initial Setup"**
section, which is common to all methods.

### 1. From Source (All Platforms)

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

2. **Prepare Initial Data:**
   On the first run, Sarracenia automatically creates a `config.json` file. However, you must manually provide the
   initial `data` directory which contains the dashboard, templates, and wordlist. Copy the contents of the `example`
   directory to where you plan to run the application.
   ```sh
   # From the repository root, copy the example contents into the current directory.
   cp -r ./example/. ./
   ```

3. **Build and Run:**
   ```sh
   # Build the application
   go build -o sarracenia ./cmd/main

   # Run the application
   ./sarracenia
   ```
   This will create an executable named `sarracenia` (or `sarracenia.exe` on Windows).

### 2. Docker (Recommended for Deployment)

This is the recommended method for deployment. Pre-built Docker images are automatically published to the GitHub Container Registry (`ghcr.io`) whenever a new version is released.

#### Running Pre-Built Images

This is the fastest and easiest way to run Sarracenia. You do not need to clone the repository.

1.  **Choose an Image Variant:**
    Two variants are available. The `native` variant is recommended for most users.
    *   **Native (Recommended):** `ghcr.io/ctag07/sarracenia:latest`
    *   **CGO:** `ghcr.io/ctag07/sarracenia:latest-cgo`

2.  **Run with `docker-compose.yml`:**
    Create a `docker-compose.yml` file with the following content.

    **For the `native` image:**
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
          - sarracenia-data:/app

    volumes:
      sarracenia-data:
    ```

3.  **Start the Service:**
    In the same directory as your `docker-compose.yml` file, run:
    ```sh
    docker compose up -d
    ```

---

#### Building from Source with Docker (for Local Development)

If you have cloned the repository and want to build the image from the source code to test local changes, you can use the provided `docker-compose.yml` file.

1.  **Choose Build Type (Optional):**
    The `docker-compose.yml` in the repository is configured to build the image. You can set the `BUILD_TYPE` environment variable.
    *   `native` (Default): Builds a pure Go binary.
    *   `cgo`: Builds using the CGO-enabled SQLite driver.

2.  **Build and Run with Docker Compose:**
    From the root of the repository:
    ```sh
    # To build and run with the default native driver
    docker compose up --build

    # To build and run with the CGO driver
    BUILD_TYPE=cgo docker compose up --build
    ```

### 3. Systemd Service (Linux)

For Linux servers, an installation script is provided to set up Sarracenia as a systemd service, ensuring it runs
automatically on boot.

**Steps:**

1. **Clone the Repository:**
   ```sh
   git clone https://github.com/CTAG07/Sarracenia.git
   cd Sarracenia
   ```

2. **Run the Installation Script:**
   The script will build the binary, create a dedicated system user (`sarracenia`), install the application files to
   `/opt/sarracenia`, and set up the systemd service.
   ```sh
   sudo bash ./scripts/install_systemd.sh
   ```

3. **Manage the Service:**
   Once installed, you can manage the Sarracenia service using standard `systemctl` commands:
    * Check status: `systemctl status sarracenia.service`
    * View logs: `journalctl -u sarracenia.service -f`
    * Stop service: `sudo systemctl stop sarracenia.service`
    * Start service: `sudo systemctl start sarracenia.service`

## Initial Setup

1. **Access the Dashboard:**
    * By default, the tarpit runs on port `:7277` and the API/Dashboard runs on port `:7278`.
    * Open a web browser and navigate to `http://localhost:7278`.

2. **Create the Master API Key:**
    * The first time you access the dashboard, the API is unsecured.
    * Navigate to the **API Keys** page.
    * Create your first key. This key will automatically be assigned the master (`*`) scope, giving it full permissions.
    * **Copy this key immediately.** It will not be shown again.
    * You will automatically be logged in with the master key. You can now tweak any part of Sarracenia to your liking.

## Components

The core components of Sarracenia are split into two standalone, reusable libraries located in the `/pkg` directory.
They have no specific dependencies on the main Sarracenia application and can be imported into other projects.

### `pkg/markov` - Go Markov Library

The core text generation engine for this project is a standalone, high-performance library for training and using Markov
chain models. It is feature-complete and includes a streaming API, database persistence, and advanced generation
features.

**[➡️ View the full README for the Markov Library](./pkg/markov/README.md)**

### `pkg/templating` - Dynamic HTML Template Engine

This library is the rendering engine for Sarracenia. It can use the text generated by the Markov library and/or
randomized text generated from a wordlist.txt, and embed it within complex, randomized HTML structures. It's designed to
be filesystem-based for easy updates and includes a rich set of template functions to generate everything from
nonsensical data and styled elements to simple quality of life operators.

Do note that this library has a dependency on the `pkg/markov` library for markov model generation, but the markov
capabilities can be disabled in the config, and all markov functions will fall back to using random words.

**[➡️ View the full README for the Templating Library](./pkg/templating/README.md)**

---

## License

This project is licensed under the AGPLv3.

**Alternative Licensing**

I understand that the AGPL-3.0 may not be suitable for all users. If you would like to use this project in a way not
permitted by the AGPL-3.0 (for instance, in a closed-source application), I am happy to grant you a free, permissive
license (such as the MIT License).

Please contact me at **`82781942+CTAG07@users.noreply.github.com`** to request an alternative license.
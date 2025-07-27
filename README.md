# Sarracenia

[![AGPLv3 License](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/CTAG07/Sarracenia)](https://goreportcard.com/report/github.com/CTAG07/Sarracenia)

A Go based anti-scraper tarpit inspired by Nepenthes.

The goal of Sarracenia is to create a hostile environment for web scrapers by generating endless, plausible-looking, and resource-intensive web pages, slowing down and polluting the data of automated crawlers.

## Project Status

This project is being built in three major phases:

- [x] **Markov Library**: A database-backed, minimal memory usage Markov chain library for text generation.
- [ ] **Templating Library**: A system for generating randomized and dynamic HTML content using the Markov library.
- [ ] **Final Tarpit Application**: The main web server and monitoring tools.

---

## Components

Components within the `/pkg` directory are designed as standalone, reusable libraries. They have no specific dependencies on the Sarracenia application and can be imported into other projects.

### `pkg/markov` - Go Markov Library

The core text generation engine for this project is a standalone, high-performance library for training and using Markov chain models. It is feature-complete and includes a streaming API, database persistence, and advanced generation features.

**[➡️ View the full README for the Markov Library](./pkg/markov/README.md)**

---

## License

This project is licensed under the AGPLv3.

**Alternative Licensing**

I understand that the AGPL-3.0 may not be suitable for all users. If you would like to use this project in a way not permitted by the AGPL-3.0 (for instance, in a closed-source application), I am happy to grant you a free, permissive license (such as the MIT License).

Please contact me at **`82781942+CTAG07@users.noreply.github.com`** to request an alternative license.

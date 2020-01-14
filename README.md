[![Go Report Card](https://goreportcard.com/badge/github.com/norgor/docker-compose-watcher)](https://goreportcard.com/report/github.com/norgor/docker-compose-watcher)
[![Build Status](https://github.com/norgor/docker-compose-watcher/workflows/build%2C%20test%20and%20release/badge.svg)](https://github.com/norgor/docker-compose-watcher/workflows/build%2C%20test%20and%20release/badge.svg)
[![Release](https://img.shields.io/github/v/release/norgor/docker-compose-watcher)](https://img.shields.io/github/v/release/norgor/docker-compose-watcher)
[![License](https://img.shields.io/github/license/norgor/docker-compose-watcher)](https://img.shields.io/github/license/norgor/docker-compose-watcher)

# About
docker-compose-watcher is an application that automatically rebuilds and restarts Docker Compose services, based on docker-compose.yaml file(s) and the specified source directories.

When the tool read the Docker Compose file, it looks for a specific label, which tells it where to look for changes. It then begins to watch both the compose files and the directory in the label for changes. A change in these files causes a rebuild and restart of the services.

# Usage
Run the docker-compose-watcher binary with the -f or --file flag(s), which specify the docker-compose files. You simply pass the same files you would pass when running docker-compose.

If you want docker-compose-watcher to watch for source directory changes, add a `docker-compose-watcher.path` label to the service (see example below).

## Example
**./repos/project/docker-compose.yml:**
~~~~~~~~~~~~~
version: "3"
services:
  my-service-name:
    build: .
    labels:
      docker-compose-watcher.path: "./"
~~~~~~~~~~~~~
**Command**
~~~~~~~~~~~~~
docker-compose-watcher -f ./repos/project/docker-compose.yml
~~~~~~~~~~~~~

## Help
Run `docker-compose-watcher --help` to print the help.

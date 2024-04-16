# Service Checker

This is a Go-based project that performs server checks and logs the results. It uses the Go standard library, and the project dependencies are managed using `go.mod`.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

You need to have Go installed on your machine. You can download it from the [official Go website](https://golang.org/dl/).

### Installing

Clone the repository to your local machine:

```bash
git clone https://github.com/ugurcsen/project.git
```

Navigate to the project directory:

```bash
cd project
```

Install the project dependencies:

```bash
go mod download
```

Build the project:

```bash
go build
```

### Usage

The project can be run using the following command:

```bash
./service-checker -c <config_file> -i <time_interval> -o <output_file> -v
```

### Example

```bash
./service-checker -c config.yaml -i 5 -o results.csv
```

Where:
- `-c <config_file>`: Specifies the configuration file to use.
- `-i <time_interval>`: Specifies the time interval for server checks.
- `-o <output_file>`: Specifies the output file to log the results.
- `-v`: Enables verbose mode.

## Project Structure

The main functionality of the project is contained within `main.go`. This file contains functions for performing server checks, decoding configuration files, and saving results.

The server checks are performed based on the configuration specified in a YAML file. The configuration includes the server's IP address, port, protocol, and other parameters.

The results of the server checks are logged and can be saved to a CSV file or sent to an OpenSearch service.

## License

This project is licensed under the Apache License - see the `LICENSE` file for details.
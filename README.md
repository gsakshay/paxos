# Paxos Implementation (Partial)
 
This system is a distributed implementation of the Paxos consensus algorithm in Go, orchestrated using Docker and Docker Compose to simulate multiple nodes. The primary goal of the system is to achieve consensus across distributed nodes, with each node playing a specific role such as a Proposer, Acceptor, or Learner.

It is built with Docker and Docker Compose, enabling the setup of multiple nodes for testing. The project includes two test cases, each managed by a separate Docker Compose file. A Makefile is provided to streamline building, running, and cleaning up the test environment.

## Prerequisites

Ensure you have the following installed on your machine:
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Makefile Commands

The Makefile in this project provides several commands to manage the Docker environment. Below is a list of commands and their descriptions.

### 1. Build the Docker Image

To build the Docker image for this project:

```bash
make build
```

This command builds a Docker image named `prj4` as specified in the Makefile. Run this command once, or after making changes to the Dockerfile.

### 2. Run Test Cases

There are two test cases defined in separate Docker Compose files:
- `docker-compose-testcase-1.yml`
- `docker-compose-testcase-2.yml`

To run each test case, use the following commands:

#### Run the First Test Case

```bash
make test1
```

This command:
- Builds the Docker image if it hasn’t been built already.
- Uses `docker-compose-testcase-1.yml` to bring up the required containers for the first test case.

#### Run the Second Test Case

```bash
make test2
```

This command:
- Builds the Docker image if needed.
- Uses `docker-compose-testcase-2.yml` to bring up the required containers for the second test case.

### 3. Stop and Remove Containers

After running a test case, use these commands to stop and remove containers for each test case:

#### Stop and Remove Containers for Test Case 1

```bash
make down-test1
```

#### Stop and Remove Containers for Test Case 2

```bash
make down-test2
```

These commands will stop the containers defined in the respective Docker Compose files, freeing up resources.

### 4. Run Both Test Cases Sequentially

To run both test cases one after the other, use:

```bash
make test
```

This command:
- Runs `make test1`, then `make down-test1` to stop containers for the first test case.
- Runs `make test2`, then `make down-test2` to stop containers for the second test case.
  
### 5. Clean Up Environment

To fully clean up the environment, removing images, unused volumes, networks, and containers:

```bash
make clean
```

This command:
- Stops any running containers for both test cases.
- Removes the Docker image (`prj4`).
- Prunes unused Docker volumes, networks, and containers to ensure a fresh start.

### 6. Display Help

To view a list of all available Makefile commands:

```bash
make help
```

This command displays the available targets in the Makefile and their descriptions.

## Notes

- Running `make test` will automatically handle both test cases sequentially, making it convenient for complete test cycles.
- The `make clean` command is powerful and will remove unused Docker resources, so use it carefully in multi-project environments.
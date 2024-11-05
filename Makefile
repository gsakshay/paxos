# Define variables
IMAGE_NAME = prj4
COMPOSE_TEST1 = docker-compose-testcase-1.yml
COMPOSE_TEST2 = docker-compose-testcase-2.yml

# Default target: build the Docker image
.PHONY: build
build:
	docker build -t $(IMAGE_NAME) .

# Run the first test case
.PHONY: test1
test1: build
	docker compose -f $(COMPOSE_TEST1) up --build

# Run the second test case
.PHONY: test2
test2: build
	docker compose -f $(COMPOSE_TEST2) up --build

# Stop and remove containers for test1
.PHONY: down-test1
down-test1:
	docker compose -f $(COMPOSE_TEST1) down

# Stop and remove containers for test2
.PHONY: down-test2
down-test2:
	docker compose -f $(COMPOSE_TEST2) down

# Clean up all containers, images, and networks
.PHONY: clean
clean: down-test1 down-test2
	docker image rm $(IMAGE_NAME) || true
	docker volume prune -f
	docker network prune -f
	docker container prune -f

# Run both test cases sequentially
.PHONY: test
test: test1 down-test1 test2 down-test2

# Display help message
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build       - Build the Docker image"
	@echo "  test1       - Run the first test case"
	@echo "  test2       - Run the second test case"
	@echo "  down-test1  - Stop and remove containers for test case 1"
	@echo "  down-test2  - Stop and remove containers for test case 2"
	@echo "  clean       - Clean up all containers, images, volumes, and networks"
	@echo "  test        - Run both test cases sequentially"
	@echo "  help        - Display this help message"

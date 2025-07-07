# Simple Makefile for a Go project

# Load variables from .env if present
# This MUST be at the top to ensure variables are available for all targets.
ifneq (,$(wildcard .env))
	include .env
	export
endif

# --- Important: Set GOOSE_DBSTRING from DATABASE_URL ---
ifdef DATABASE_URL
export GOOSE_DBSTRING = $(DATABASE_URL)
endif

# --- Optional: Set GOOSE_DRIVER if not already in .env or you want to be explicit ---
ifdef GOOSE_DRIVER
# Already defined in .env, no action needed
else
export GOOSE_DRIVER = postgres
endif

# Define the Goose command.
# Goose will automatically pick up GOOSE_DRIVER and GOOSE_DBSTRING from the environment.
GOOSE = goose -dir migrations

# ==============================================================================
# General Build and Run Targets
# ==============================================================================

all: build test

templ-install:
	@if ! command -v templ > /dev/null; then \
		read -p "Go's 'templ' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/a-h/templ/cmd/templ@latest; \
			if [ ! -x "$$(command -v templ)" ]; then \
				echo "templ installation failed. Exiting..."; \
				exit 1; \
			fi; \
		else \
			echo "You chose not to install templ. Exiting..."; \
			exit 1; \
		fi; \
	fi
tailwind-install:
	@if [ ! -f tailwindcss ]; then curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 -o tailwindcss; fi
	
	@chmod +x tailwindcss

build: tailwind-install templ-install
	@echo "Building..."
	@templ generate
	@./tailwindcss -i cmd/web/styles/input.css -o cmd/web/assets/css/output.css
	@CGO_ENABLED=1 GOOS=linux go build -o main cmd/api/main.go

run:
	@go run cmd/api/main.go

docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

test:
	@echo "Testing..."
	@go test ./... -v

clean:
	@echo "Cleaning..."
	@rm -f main

watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# ==============================================================================
# Goose Migration Targets
# ==============================================================================

# For debugging purposes: print the database URL Goose will use
print-db:
	@echo "DATABASE_URL (from .env): $(DATABASE_URL)"
	@echo "GOOSE_DBSTRING (for Goose): $(GOOSE_DBSTRING)"
	@echo "GOOSE_DRIVER: $(GOOSE_DRIVER)"
	@echo "Migrations directory: migrations"

migrate-up:
	$(GOOSE) up

migrate-down:
	$(GOOSE) down

migrate-status:
	$(GOOSE) status

migrate-create:
ifeq ($(strip $(name)),)
	$(error name is required: make migrate-create name=create_users_table type=sql)
endif
ifeq ($(strip $(type)),)
	$(error type is required: make migrate-create name=create_users_table type=sql or type=go)
endif
	$(GOOSE) create $(name) $(type)

migrate-reset:
	@echo "⚠️ WARNING: This will erase and recreate your database!"
	@bash -c 'read -p "Are you sure you want to reset the DB? [y/N] " confirm && \
	if [ "$$confirm" = "y" ]; then \
		GOOSE_DRIVER="$(GOOSE_DRIVER)" GOOSE_DBSTRING="$(GOOSE_DBSTRING)" $(GOOSE) reset; \
	else \
		echo "❌ Reset canceled."; \
	fi'

.PHONY: all build run test clean watch tailwind-install templ-install print-db migrate-up migrate-down migrate-status migrate-create migrate-reset

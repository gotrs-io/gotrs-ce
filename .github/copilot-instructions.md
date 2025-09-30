# GitHub Copilot Instructions for GOTRS

## Project Context
Building GOTRS - a modern ticketing system in Go and React.

## Code Style
- Minimal comments
- Direct implementation
- Assume imports exist
- Follow existing patterns

## Avoid
- Long explanations
- Decorative comments
- Unnecessary abstractions
- Verbose error handling

## Focus
- Working code first try
- Efficient solutions
- Reuse existing code
- Simple error returns

## Lessons Learned (Container-First Development)
- **Running the app**: Use `make up` to start it, and `make down` to stop it. `make restart` **builds** and restarts the app.
- **Always use Makefile targets**: Never run direct commands on host (e.g., `go`, `mysql`, `npm`). Use `make toolbox-*`, `make db-*`, `make css-*` instead, these targets automatically assume the correct credentials too.
- **Database operations**: Use `make db-shell`, `make db-query`, `make db-migrate` - never direct `mysql` commands.
- **Go operations**: Use `make toolbox-exec ARGS="go <command>"` - never run `go` directly on host.
- **Container-first enforcement**: All development, testing, and operations must go through containers via make targets.
- **Avoid anti-patterns**: Don't create host dependencies; assume Go, Node.js, and database clients are not installed locally.
- **Testing in containers**: All tests run via `make test-*` targets using toolbox containers.
- **Security and consistency**: Containerized operations ensure proper permissions, caching, and environment isolation.
- **Command generation**: When suggesting commands to run, always use make targets and container operations - never suggest direct tool execution on host.
- **No direct docker/podman commands**: Even in examples or instructions, use `make` targets instead of `docker compose run` or similar.
- **Toolbox for everything**: All Go, database, and development operations must route through the toolbox container via make targets.
</markdown># GitHub Copilot Instructions for GOTRS

## Project Context
Building GOTRS - a modern ticketing system in Go and React.

## Code Style
- Minimal comments
- Direct implementation
- Assume imports exist
- Follow existing patterns

## Avoid
- Long explanations
- Decorative comments
- Unnecessary abstractions
- Verbose error handling
- '!' symbol in bash commands
- creating temporary files anywhere except in tmp/ which is ignored by git

## Focus
- Working code first try
- Efficient solutions
- Reuse existing code
- Simple error returns

## Lessons Learned (Container-First Development)
- **Running the app**: Use `make up` to start it, and `make down` to stop it. `make restart` **builds** and restarts the app.
- **Always use Makefile targets**: Never run direct commands on host (e.g., `go`, `mysql`, `npm`). Use `make toolbox-*`, `make db-*`, `make css-*` instead, these targets automatically assume the correct credentials too.
- **Database operations**: Use `make db-shell`, `make db-query`, `make db-migrate` - never direct `mysql` commands.
- **Go operations**: Use `make toolbox-exec ARGS="go <command>"` - never run `go` directly on host.
- **Container-first enforcement**: All development, testing, and operations must go through containers via make targets.
- **Avoid anti-patterns**: Don't create host dependencies; assume Go, Node.js, and database clients are not installed locally.
- **Testing in containers**: All tests run via `make test-*` targets using toolbox containers.
- **Security and consistency**: Containerized operations ensure proper permissions, caching, and environment isolation.
- **Command generation**: When suggesting commands to run, always use make targets and container operations - never suggest direct tool execution on host.
- **No direct docker/podman commands**: Even in examples or instructions, use `make` targets instead of `docker compose run` or similar.
- **Toolbox for everything**: All Go, database, and development operations must route through the toolbox container via make targets.
</markdown>

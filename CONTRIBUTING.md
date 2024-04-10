# Setup

- Make sure you have docker container runtime installed.

# Running

- Create a *.env* file following *.env.example* and populate the variables.
- Run the containers for development using `docker compose -f compose.dev.yml up -d`.
- Run `make migrate` to import the schema defined at *tools/schema.surql*.
- Run the respective cron jobs using `make cron task={{file_name}}`
- The server will start on [`http://localhost:3000`](http://localhost:3000) using [air](https://github.com/cosmtrek/air) in a docker container for development with hot reload.

## Exploring API

- Download and setup the [Bruno API client](https://www.usebruno.com/).
- The collection is stored in the *api* folder.
- Choose the *Local* environment and setup your creds for testing.
- Explore the API (PS: you can view the docs of individual requests in bruno).

## Testing

- Modify the *.env.testing* file
- Run `make test`
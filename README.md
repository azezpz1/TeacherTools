# Teacher Tools

## Setup

1. Create your own .env file based on sample.env.
2. Run using docker compose instead of running the Dockerfile or the go application directly

## Creating an account

Send a POST request to `/signup` with JSON body:

```json
{
  "email": "you@example.com",
  "password": "yourpassword"
}
```

On success the server responds with HTTP 201.

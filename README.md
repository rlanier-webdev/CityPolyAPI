# Rivalry API v2
------------
## Introduction
------------
This API provides endpoints to retrieve information about the historic City-Poly rivalry game.

Version 2 introduces user accounts and API key authentication. All data endpoints require an API key passed via the `X-API-Key` header.

## Try it in Postman
------------
[![Run in Postman](https://run.pstmn.io/button.svg)](https://www.postman.com/thrtn85/main/api/2451bd7d-f19d-4b66-9317-a605ced2ba0c/entity?action=share&creator=13028315)

---

## Getting Started
------------
To access data endpoints you need an API key. Here's how to get one:

1. **Register** — Create an account with your email and password.
2. **Login** — Receive a bearer token.
3. **Create API Key** — Use the bearer token to generate an API key.
4. **Use the API Key** — Pass it as `X-API-Key` on all data requests.

---

## Authentication Endpoints
------------
### Register
- **URL**: `/api/auth/register`
- **Method**: `POST`
- **Auth**: None
- **Body**:
  ```json
  { "email": "you@example.com", "password": "YourPassword1!" }
  ```
- **Response**:
  - `201 Created` — Account created.
  - `400 Bad Request` — Invalid input.
  - `409 Conflict` — Email already registered.

---

### Login
- **URL**: `/api/auth/login`
- **Method**: `POST`
- **Auth**: None
- **Body**:
  ```json
  { "email": "you@example.com", "password": "YourPassword1!" }
  ```
- **Response**:
  - `200 OK` — Returns a bearer token. Store it — you'll need it to create API keys.
  ```json
  { "token": "bearer_..." }
  ```
  - `401 Unauthorized` — Invalid credentials.

---

### Create API Key
- **URL**: `/api/auth/keys`
- **Method**: `POST`
- **Auth**: `Authorization: Bearer <token>`
- **Response**:
  - `201 Created` — Returns your API key. **Store it — it will not be shown again.**
  ```json
  { "key": "riv_...", "note": "Store this safely. It will not be shown again." }
  ```

---

### List API Keys
- **URL**: `/api/auth/keys`
- **Method**: `GET`
- **Auth**: `Authorization: Bearer <token>`
- **Response**:
  - `200 OK` — Returns a list of your active API keys (prefixes only, not full keys).

---

### Revoke API Key
- **URL**: `/api/auth/keys/:id`
- **Method**: `DELETE`
- **Auth**: `Authorization: Bearer <token>`
- **Response**:
  - `204 No Content` — Key revoked.
  - `404 Not Found` — Key not found or doesn't belong to your account.

---

### Logout
- **URL**: `/api/auth/logout`
- **Method**: `POST`
- **Auth**: `Authorization: Bearer <token>`
- **Response**:
  - `204 No Content` — Bearer token revoked.

---

## Data Endpoints
------------
All data endpoints require an API key:
```
X-API-Key: riv_...
```

### Get All Games
- **URL**: `/api/v2/games`
- **Method**: `GET`
- **Description**: Returns all games in JSON format.
- **Response**:
  - `200 OK` — JSON array of games.
  - `401 Unauthorized` — Missing or invalid API key.

### Get Game by ID
- **URL**: `/api/v2/games/:id`
- **Method**: `GET`
- **Description**: Returns the game with the specified ID.
- **Parameters**:
  - `id` (int): The ID of the game to retrieve.
- **Response**:
  - `200 OK` — JSON object representing the game.
  - `404 Not Found` — Game not found.

### Get Games by Home Team
- **URL**: `/api/v2/games/home/:name`
- **Method**: `GET`
- **Description**: Returns all games where the specified team played as the home team.
- **Parameters**:
  - `name` (string): Name of the home team.
- **Response**:
  - `200 OK` — JSON array of matching games.
  - `404 Not Found` — No games found.

### Get Games by Away Team
- **URL**: `/api/v2/games/away/:name`
- **Method**: `GET`
- **Description**: Returns all games where the specified team played as the away team.
- **Parameters**:
  - `name` (string): Name of the away team.
- **Response**:
  - `200 OK` — JSON array of matching games.
  - `404 Not Found` — No games found.

### Get Games by Year
- **URL**: `/api/v2/games/year/:year`
- **Method**: `GET`
- **Description**: Returns all games played in the specified year.
- **Parameters**:
  - `year` (int): The year to retrieve games for.
- **Response**:
  - `200 OK` — JSON array of games.
  - `404 Not Found` — No games found for that year.

### Get All Teams
- **URL**: `/api/v2/teams`
- **Method**: `GET`
- **Description**: Returns all unique team names.
- **Response**:
  - `200 OK` — JSON object with teams array.

---

## Data Structure
------------
### Game
- `id` (int): Unique identifier.
- `home_team` (string): The home team.
- `away_team` (string): The away team.
- `date` (string): Date of the game (`YYYY-MM-DD`).
- `home_team_score` (int): Home team score.
- `away_team_score` (int): Away team score.
- `notes` (string): Additional notes.

---

## Rate Limiting
------------
Requests are rate limited to **10 per second** per IP address with a burst allowance of 20. Exceeding this returns `429 Too Many Requests`.

# LocationAgent

A location sharing service with Firebase phone authentication.

## Setup

1. Install dependencies:

```bash
go mod download
```

2. Configure Firebase:

- Place your `serviceAccountKey.json` in the `config` directory
- Set your Firebase Web API Key in environment:

```bash
export FIREBASE_API_KEY=your_api_key
```

3. Run tests:

```bash
go test -v ./src/test/integration_test.go
```

## Project Structure

```
agent-backend
├── src
│   ├── main.go          # Entry point of the application
│   ├── handlers         # Contains HTTP request handlers
│   │   └── handler.go   # Functions to handle incoming requests
│   ├── routes           # Defines application routes
│   │   └── routes.go    # Setup routes for the application
│   └── models           # Data structures for the application
│       └── model.go     # Defines the Item struct
├── go.mod               # Module definition and dependencies
└── README.md            # Project documentation
```

## Usage

- **Get Items:**

  - Endpoint: `GET /items`
  - Description: Retrieves a list of items.

- **Create Item:**
  - Endpoint: `POST /items`
  - Description: Creates a new item. Requires a JSON body with item details.

## Contributing

Feel free to submit issues or pull requests for improvements or bug fixes.

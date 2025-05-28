# GEAV Site API

This project provides a serverless API for the GEAV site, implementing CRUD operations for users, places (lugares), and songs (cancoes).

## Features

- Serverless architecture using AWS Lambda
- PostgreSQL database for data storage
- CloudWatch logging for API actions
- Database logging for API actions
- Infrastructure as Code using AWS CloudFormation

## Project Structure

- `cmd/`: Contains the main applications for the project
- `internal/`: Contains private application and library code
  - `models/`: Database models
  - `handlers/`: Lambda function handlers
  - `repository/`: Database access layer
  - `logger/`: Logging functionality
- `pkg/`: Contains code that's ok for other services to consume
- `infrastructure/`: Contains CloudFormation templates
- `scripts/`: Contains utility scripts

## Requirements

- Go 1.21 or higher
- AWS CLI
- AWS SAM CLI (for local testing)
- PostgreSQL (for local development)

## Setup

1. Clone the repository
2. Run `go mod tidy` to download dependencies
3. Configure AWS credentials using the AWS CLI:
   ```
   aws configure
   ```
4. Deploy the application using the deployment script:
   ```
   .\scripts\deploy.ps1 -StackName geav-site-api -DBUsername admin -DBPassword your-password
   ```
5. Initialize the database using the command provided by the deployment script

### Deployment Parameters

The deployment script accepts the following parameters:

- `StackName` (required): The name of the CloudFormation stack
- `Environment` (optional, default: "dev"): The environment to deploy to (dev or prod)
- `DBUsername` (required): The username for the PostgreSQL database
- `DBPassword` (required): The password for the PostgreSQL database
- `DBName` (optional, default: "geav"): The name of the PostgreSQL database
- `Region` (optional, default: "us-east-1"): The AWS region to deploy to

## Local Testing

You can test the API locally using the AWS SAM CLI before deploying it to AWS. This allows you to verify that your changes work as expected.

### Prerequisites

1. Install the AWS SAM CLI:
   - Follow the instructions at [AWS SAM CLI Installation Guide](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)

2. Set up a local PostgreSQL database:
   - Create a database named `geav`
   - Initialize the database using the schema in `scripts/init-db.sql`

### Running Tests Locally

Use the provided script to test the API locally:

```
.\scripts\test-local.ps1 -Function users -Event get-all -DBHost localhost -DBPort 5432 -DBUsername postgres -DBPassword postgres -DBName geav
```

### Local Testing Parameters

The local testing script accepts the following parameters:

- `Function` (optional, default: "users"): The Lambda function to test (users, lugares, or cancoes)
- `Event` (optional, default: "get-all"): The event to use for testing (get-all, get-one, create, update, or delete)
- `DBHost` (optional, default: "localhost"): The hostname of the PostgreSQL database
- `DBPort` (optional, default: "5432"): The port of the PostgreSQL database
- `DBUsername` (optional, default: "postgres"): The username for the PostgreSQL database
- `DBPassword` (optional, default: "postgres"): The password for the PostgreSQL database
- `DBName` (optional, default: "geav"): The name of the PostgreSQL database

## API Endpoints

The API provides the following endpoints:

### Users
- `GET /users`: List all users
- `GET /users/{id}`: Get a specific user
- `POST /users`: Create a new user
- `PUT /users/{id}`: Update a user
- `DELETE /users/{id}`: Delete a user

### Places (Lugares)
- `GET /lugares`: List all places
- `GET /lugares/{id}`: Get a specific place
- `POST /lugares`: Create a new place
- `PUT /lugares/{id}`: Update a place
- `DELETE /lugares/{id}`: Delete a place

### Songs (Cancoes)
- `GET /cancoes`: List all songs
- `GET /cancoes/{id}`: Get a specific song
- `POST /cancoes`: Create a new song
- `PUT /cancoes/{id}`: Update a song
- `DELETE /cancoes/{id}`: Delete a song

## License

This project is licensed under the MIT License - see the LICENSE file for details.

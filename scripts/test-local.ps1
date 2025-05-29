# Local testing script for GEAV Site API
# This script builds the Lambda functions and runs them locally using AWS SAM CLI

param (
    [Parameter(Mandatory=$false)]
    [string]$Function = "users",

    [Parameter(Mandatory=$false)]
    [string]$Event = "get-all",

    [Parameter(Mandatory=$false)]
    [string]$DBHost = "localhost",

    [Parameter(Mandatory=$false)]
    [string]$DBPort = "5432",

    [Parameter(Mandatory=$false)]
    [string]$DBUsername = "postgres",

    [Parameter(Mandatory=$false)]
    [string]$DBPassword = "pgadmin",

    [Parameter(Mandatory=$false)]
    [string]$DBName = "geav"
)

# Check if AWS SAM CLI is installed
$samInstalled = Get-Command sam -ErrorAction SilentlyContinue
if (-not $samInstalled) {
    Write-Host "AWS SAM CLI is not installed. Please install it first." -ForegroundColor Red
    Write-Host "https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html" -ForegroundColor Yellow
    exit 1
}

# Create build directory
$buildDir = ".\build"
if (-not (Test-Path $buildDir)) {
    New-Item -Path $buildDir -ItemType Directory | Out-Null
}

# Create events directory
$eventsDir = ".\events"
if (-not (Test-Path $eventsDir)) {
    New-Item -Path $eventsDir -ItemType Directory | Out-Null
}

# Create sample events if they don't exist
if (-not (Test-Path "$eventsDir\users-get-all.json")) {
    @"
{
  "resource": "/users",
  "path": "/users",
  "httpMethod": "GET",
  "headers": {
    "Accept": "*/*",
    "x-request-id": "test-request-id"
  },
  "queryStringParameters": null,
  "pathParameters": null,
  "requestContext": {
    "resourceId": "123456",
    "resourcePath": "/users",
    "httpMethod": "GET",
    "requestId": "test-request-id"
  },
  "body": null,
  "isBase64Encoded": false
}
"@ | Out-File -FilePath "$eventsDir\users-get-all.json" -Encoding utf8
}

if (-not (Test-Path "$eventsDir\users-get-one.json")) {
    @"
{
  "resource": "/users/{id}",
  "path": "/users/1",
  "httpMethod": "GET",
  "headers": {
    "Accept": "*/*",
    "x-request-id": "test-request-id"
  },
  "queryStringParameters": null,
  "pathParameters": {
    "id": "1"
  },
  "requestContext": {
    "resourceId": "123456",
    "resourcePath": "/users/{id}",
    "httpMethod": "GET",
    "requestId": "test-request-id"
  },
  "body": null,
  "isBase64Encoded": false
}
"@ | Out-File -FilePath "$eventsDir\users-get-one.json" -Encoding utf8
}

if (-not (Test-Path "$eventsDir\users-create.json")) {
    @"
{
  "resource": "/users",
  "path": "/users",
  "httpMethod": "POST",
  "headers": {
    "Accept": "*/*",
    "Content-Type": "application/json",
    "x-request-id": "test-request-id"
  },
  "queryStringParameters": null,
  "pathParameters": null,
  "requestContext": {
    "resourceId": "123456",
    "resourcePath": "/users",
    "httpMethod": "POST",
    "requestId": "test-request-id"
  },
  "body": "{\"username\":\"testuser\",\"password\":\"testpassword\",\"role\":\"read\"}",
  "isBase64Encoded": false
}
"@ | Out-File -FilePath "$eventsDir\users-create.json" -Encoding utf8
}

if (-not (Test-Path "$eventsDir\users-update.json")) {
    @"
{
  "resource": "/users/{id}",
  "path": "/users/1",
  "httpMethod": "PUT",
  "headers": {
    "Accept": "*/*",
    "Content-Type": "application/json",
    "x-request-id": "test-request-id"
  },
  "queryStringParameters": null,
  "pathParameters": {
    "id": "1"
  },
  "requestContext": {
    "resourceId": "123456",
    "resourcePath": "/users/{id}",
    "httpMethod": "PUT",
    "requestId": "test-request-id"
  },
  "body": "{\"username\":\"updateduser\",\"password\":\"updatedpassword\",\"role\":\"write\"}",
  "isBase64Encoded": false
}
"@ | Out-File -FilePath "$eventsDir\users-update.json" -Encoding utf8
}

if (-not (Test-Path "$eventsDir\users-delete.json")) {
    @"
{
  "resource": "/users/{id}",
  "path": "/users/1",
  "httpMethod": "DELETE",
  "headers": {
    "Accept": "*/*",
    "x-request-id": "test-request-id"
  },
  "queryStringParameters": null,
  "pathParameters": {
    "id": "1"
  },
  "requestContext": {
    "resourceId": "123456",
    "resourcePath": "/users/{id}",
    "httpMethod": "DELETE",
    "requestId": "test-request-id"
  },
  "body": null,
  "isBase64Encoded": false
}
"@ | Out-File -FilePath "$eventsDir\users-delete.json" -Encoding utf8
}

# Build the Lambda function
Write-Host "Building $Function Lambda function..." -ForegroundColor Green
go build -o $buildDir\$Function .\cmd\$Function\main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build $Function Lambda function" -ForegroundColor Red
    exit 1
}

# Determine the event file
$eventFile = "$eventsDir\$Function-$Event.json"
if (-not (Test-Path $eventFile)) {
    Write-Host "Event file $eventFile does not exist" -ForegroundColor Red
    exit 1
}

# Run the Lambda function locally
Write-Host "Running $Function Lambda function locally..." -ForegroundColor Green
$env:DB_HOST = $DBHost
$env:DB_PORT = $DBPort
$env:DB_USER = $DBUsername
$env:DB_PASSWORD = $DBPassword
$env:DB_NAME = $DBName
$env:ENVIRONMENT = "local"

sam local invoke $Function"Function" `
    --event $eventFile `
    --parameter-overrides "ParameterKey=Environment,ParameterValue=local ParameterKey=DBHost,ParameterValue=$DBHost ParameterKey=DBPort,ParameterValue=$DBPort ParameterKey=DBUsername,ParameterValue=$DBUsername ParameterKey=DBPassword,ParameterValue=$DBPassword ParameterKey=DBName,ParameterValue=$DBName" `
    --template template.yaml `
    --debug

Write-Host "Local testing completed!" -ForegroundColor Green

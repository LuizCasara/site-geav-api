# Deployment script for GEAV Site API
# This script builds the Lambda functions, uploads them to S3, and deploys the CloudFormation stack

param (
    [Parameter(Mandatory=$true)]
    [string]$StackName,
    
    [Parameter(Mandatory=$false)]
    [string]$Environment = "dev",
    
    [Parameter(Mandatory=$true)]
    [string]$DBUsername,
    
    [Parameter(Mandatory=$true)]
    [string]$DBPassword,
    
    [Parameter(Mandatory=$false)]
    [string]$DBName = "geav",
    
    [Parameter(Mandatory=$false)]
    [string]$Region = "us-east-1"
)

# Set AWS region
$env:AWS_REGION = $Region

# Create build directory
$buildDir = ".\build"
if (Test-Path $buildDir) {
    Remove-Item -Path $buildDir -Recurse -Force
}
New-Item -Path $buildDir -ItemType Directory | Out-Null

# Build Lambda functions
Write-Host "Building Lambda functions..." -ForegroundColor Green

# Build users Lambda function
Write-Host "Building users Lambda function..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o $buildDir\users .\cmd\users\main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build users Lambda function" -ForegroundColor Red
    exit 1
}

# Build lugares Lambda function
Write-Host "Building lugares Lambda function..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o $buildDir\lugares .\cmd\lugares\main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build lugares Lambda function" -ForegroundColor Red
    exit 1
}

# Build cancoes Lambda function
Write-Host "Building cancoes Lambda function..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o $buildDir\cancoes .\cmd\cancoes\main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build cancoes Lambda function" -ForegroundColor Red
    exit 1
}

# Create zip files for Lambda functions
Write-Host "Creating zip files for Lambda functions..." -ForegroundColor Green

# Create users.zip
Write-Host "Creating users.zip..." -ForegroundColor Yellow
Compress-Archive -Path $buildDir\users -DestinationPath $buildDir\users.zip -Force

# Create lugares.zip
Write-Host "Creating lugares.zip..." -ForegroundColor Yellow
Compress-Archive -Path $buildDir\lugares -DestinationPath $buildDir\lugares.zip -Force

# Create cancoes.zip
Write-Host "Creating cancoes.zip..." -ForegroundColor Yellow
Compress-Archive -Path $buildDir\cancoes -DestinationPath $buildDir\cancoes.zip -Force

# Create S3 bucket for Lambda code
$s3BucketName = "$StackName-lambda-code-$(aws sts get-caller-identity --query 'Account' --output text)"
Write-Host "Creating S3 bucket $s3BucketName..." -ForegroundColor Green

# Check if bucket exists
$bucketExists = aws s3api head-bucket --bucket $s3BucketName 2>&1
if ($LASTEXITCODE -ne 0) {
    # Create bucket
    aws s3 mb s3://$s3BucketName --region $Region
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Failed to create S3 bucket" -ForegroundColor Red
        exit 1
    }
}

# Upload Lambda code to S3
Write-Host "Uploading Lambda code to S3..." -ForegroundColor Green

# Upload users.zip
Write-Host "Uploading users.zip..." -ForegroundColor Yellow
aws s3 cp $buildDir\users.zip s3://$s3BucketName/users.zip
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to upload users.zip to S3" -ForegroundColor Red
    exit 1
}

# Upload lugares.zip
Write-Host "Uploading lugares.zip..." -ForegroundColor Yellow
aws s3 cp $buildDir\lugares.zip s3://$s3BucketName/lugares.zip
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to upload lugares.zip to S3" -ForegroundColor Red
    exit 1
}

# Upload cancoes.zip
Write-Host "Uploading cancoes.zip..." -ForegroundColor Yellow
aws s3 cp $buildDir\cancoes.zip s3://$s3BucketName/cancoes.zip
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to upload cancoes.zip to S3" -ForegroundColor Red
    exit 1
}

# Deploy CloudFormation stack
Write-Host "Deploying CloudFormation stack..." -ForegroundColor Green
aws cloudformation deploy `
    --template-file .\infrastructure\cloudformation.yaml `
    --stack-name $StackName `
    --parameter-overrides `
        Environment=$Environment `
        DBUsername=$DBUsername `
        DBPassword=$DBPassword `
        DBName=$DBName `
    --capabilities CAPABILITY_IAM `
    --region $Region

if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to deploy CloudFormation stack" -ForegroundColor Red
    exit 1
}

# Get stack outputs
Write-Host "Getting stack outputs..." -ForegroundColor Green
$outputs = aws cloudformation describe-stacks --stack-name $StackName --query "Stacks[0].Outputs" --output json | ConvertFrom-Json

# Display API endpoint
$apiEndpoint = ($outputs | Where-Object { $_.OutputKey -eq "ApiEndpoint" }).OutputValue
Write-Host "API Endpoint: $apiEndpoint" -ForegroundColor Cyan

# Display database endpoint
$dbEndpoint = ($outputs | Where-Object { $_.OutputKey -eq "DatabaseEndpoint" }).OutputValue
$dbPort = ($outputs | Where-Object { $_.OutputKey -eq "DatabasePort" }).OutputValue
Write-Host "Database Endpoint: $dbEndpoint" -ForegroundColor Cyan
Write-Host "Database Port: $dbPort" -ForegroundColor Cyan

# Initialize database
Write-Host "Initializing database..." -ForegroundColor Green
Write-Host "To initialize the database, run the following command:" -ForegroundColor Yellow
Write-Host "psql -h $dbEndpoint -p $dbPort -U $DBUsername -d $DBName -f .\scripts\init-db.sql" -ForegroundColor Yellow

Write-Host "Deployment completed successfully!" -ForegroundColor Green
@echo off
REM Tanzanite Go Backend Startup Script for Windows

echo Starting Tanzanite Go Backend...

REM Check if config file exists
if not exist "config\config.yaml" (
    echo Config file not found. Copying from example...
    copy config\config.example.yaml config\config.yaml
    echo Config file created. Please update config\config.yaml with your settings.
)

REM Check if .env file exists
if not exist ".env" (
    echo .env file not found. Copying from example...
    copy .env.example .env
    echo .env file created. Please update .env with your settings.
)

REM Download dependencies
echo Downloading dependencies...
go mod download

REM Build the application
echo Building application...
go build -o tanzanite-api.exe .\cmd\server

REM Run the application
echo Starting server...
tanzanite-api.exe

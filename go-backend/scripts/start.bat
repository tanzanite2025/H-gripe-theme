@echo off
setlocal EnableExtensions
REM Storefront Go Backend Startup Script for Windows

for %%I in ("%~dp0..") do set "ROOT_DIR=%%~fI"
set "BIN_DIR=%ROOT_DIR%\bin"
set "BIN_BASE=%BIN_DIR%\server"
set "BINARY=%BIN_BASE%"

echo Starting Storefront Go Backend...

pushd "%ROOT_DIR%" >nul

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
if errorlevel 1 goto :error

REM Build the application into the shared bin directory
echo Building application...
if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"
go build -o "%BIN_BASE%" .\cmd\server
if errorlevel 1 goto :error

if exist "%BIN_BASE%.exe" set "BINARY=%BIN_BASE%.exe"

REM Run the application
echo Starting server...
"%BINARY%"
set "EXIT_CODE=%ERRORLEVEL%"
goto :cleanup

:error
set "EXIT_CODE=%ERRORLEVEL%"
echo Failed to start Storefront Go Backend.

:cleanup
popd >nul
exit /b %EXIT_CODE%

@echo off
setlocal EnableExtensions EnableDelayedExpansion

set "OUTPUT_DIR=dist"
set "BINARY_NAME=archiver"
set "ARCH=amd64"
set "CLEAN=0"

:parse_args
if "%~1"=="" goto args_done
if /I "%~1"=="--out-dir" (
  if "%~2"=="" goto arg_error
  set "OUTPUT_DIR=%~2"
  shift
  shift
  goto parse_args
)
if /I "%~1"=="--name" (
  if "%~2"=="" goto arg_error
  set "BINARY_NAME=%~2"
  shift
  shift
  goto parse_args
)
if /I "%~1"=="--arch" (
  if "%~2"=="" goto arg_error
  set "ARCH=%~2"
  shift
  shift
  goto parse_args
)
if /I "%~1"=="--clean" (
  set "CLEAN=1"
  shift
  goto parse_args
)
if /I "%~1"=="-h" goto usage
if /I "%~1"=="--help" goto usage

echo Unknown argument: %~1

goto usage

:arg_error
echo Missing value for argument: %~1

goto usage

:args_done
if /I not "%ARCH%"=="amd64" if /I not "%ARCH%"=="arm64" (
  echo Invalid --arch value: %ARCH%
  echo Supported values: amd64, arm64
  exit /b 1
)

where go >nul 2>nul
if errorlevel 1 (
  echo Go is not installed or not in PATH.
  exit /b 1
)

set "SCRIPT_DIR=%~dp0"
for %%I in ("%SCRIPT_DIR%..") do set "ROOT_DIR=%%~fI"
pushd "%ROOT_DIR%" >nul
if errorlevel 1 (
  echo Failed to enter project directory: %ROOT_DIR%
  exit /b 1
)

if "%CLEAN%"=="1" if exist "%OUTPUT_DIR%" (
  rmdir /s /q "%OUTPUT_DIR%"
)
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

set "TARGET=%OUTPUT_DIR%\%BINARY_NAME%-windows-%ARCH%.exe"

echo Building Windows executable...
echo   Project : %ROOT_DIR%
echo   Target  : %TARGET%

set "CGO_ENABLED=0"
set "GOOS=windows"
set "GOARCH=%ARCH%"

go build -trimpath -ldflags "-s -w" -o "%TARGET%" .
if errorlevel 1 (
  echo go build failed.
  popd >nul
  exit /b 1
)

echo Build completed: %TARGET%
popd >nul
exit /b 0

:usage
echo Usage:
echo   scripts\build_windows.bat [--out-dir DIR] [--name NAME] [--arch amd64^|arm64] [--clean]
echo.
echo Examples:
echo   scripts\build_windows.bat
echo   scripts\build_windows.bat --arch arm64
echo   scripts\build_windows.bat --out-dir release --name kopi --clean
exit /b 1

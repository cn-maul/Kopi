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

echo 未知参数: %~1

goto usage

:arg_error
echo 参数缺少值: %~1

goto usage

:args_done
if /I not "%ARCH%"=="amd64" if /I not "%ARCH%"=="arm64" (
  echo --arch 参数无效: %ARCH%
  echo 支持的取值: amd64, arm64
  exit /b 1
)

where go >nul 2>nul
if errorlevel 1 (
  echo 未检测到 Go，或 Go 不在 PATH 中。
  exit /b 1
)

set "SCRIPT_DIR=%~dp0"
for %%I in ("%SCRIPT_DIR%..") do set "ROOT_DIR=%%~fI"
pushd "%ROOT_DIR%" >nul
if errorlevel 1 (
  echo 进入项目目录失败: %ROOT_DIR%
  exit /b 1
)

if "%CLEAN%"=="1" if exist "%OUTPUT_DIR%" (
  rmdir /s /q "%OUTPUT_DIR%"
)
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

set "TARGET=%OUTPUT_DIR%\%BINARY_NAME%-windows-%ARCH%.exe"

echo 正在编译 Windows 可执行文件...
echo   项目目录: %ROOT_DIR%
echo   输出文件: %TARGET%

set "CGO_ENABLED=0"
set "GOOS=windows"
set "GOARCH=%ARCH%"

go build -trimpath -ldflags "-s -w" -o "%TARGET%" .
if errorlevel 1 (
  echo go build 编译失败。
  popd >nul
  exit /b 1
)

echo 编译完成: %TARGET%
popd >nul
exit /b 0

:usage
echo 用法:
echo   scripts\build_windows.bat [--out-dir 目录] [--name 名称] [--arch amd64^|arm64] [--clean]
echo.
echo 示例:
echo   scripts\build_windows.bat
echo   scripts\build_windows.bat --arch arm64
echo   scripts\build_windows.bat --out-dir release --name kopi --clean
exit /b 1

@echo off
setlocal
cd /d "%~dp0"

set APP=ramcleaner.exe

where go >nul 2>nul
if errorlevel 1 (
  echo [fail] go not found in PATH
  pause
  exit /b 1
)

where garble >nul 2>nul
if errorlevel 1 (
  echo [fail] garble not found in PATH
  pause
  exit /b 1
)

if exist "%APP%" del /f /q "%APP%"

echo [*] building %APP% with garble...
garble -literals -seed=random build -trimpath -ldflags="-s -w" -o "%APP%" .
if errorlevel 1 (
  echo [fail] build failed
  pause
  exit /b 1
)

echo [ok] build done: %APP%
endlocal

@echo off
REM For when the admin passphrase is lost. Stops the service (if running),
REM clears it, prints a fresh one, then restarts the service. Safe to run
REM even if WaqtiService isn't installed/running -- the stop/start calls
REM are best-effort. Does not touch anything else in data\.

set SERVICE_NAME=WaqtiService
set INSTALL_DIR=%~dp0..
set EXE_PATH=%INSTALL_DIR%\waqti.exe

if not exist "%EXE_PATH%" (
    echo ERROR: %EXE_PATH% not found.
    pause
    exit /b 1
)

echo Stopping %SERVICE_NAME% if it's running...
nssm.exe stop %SERVICE_NAME% >nul 2>&1

echo.
"%EXE_PATH%" --reset-passphrase
echo.

echo Restarting %SERVICE_NAME% if it's installed...
nssm.exe start %SERVICE_NAME% >nul 2>&1

echo Done. Copy the passphrase printed above, then log in at
echo http://localhost:3000/admin and change it from the settings page.
pause

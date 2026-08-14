@echo off
REM Registers waqti as a Windows background service via NSSM.
REM Requires nssm.exe on PATH (https://nssm.cc/). Run as Administrator.

set SERVICE_NAME=WaqtiService
set INSTALL_DIR=%~dp0..
set EXE_PATH=%INSTALL_DIR%\waqti.exe

if not exist "%EXE_PATH%" (
    echo ERROR: %EXE_PATH% not found. Build the binary first ^(make build-windows^).
    exit /b 1
)

nssm.exe install %SERVICE_NAME% "%EXE_PATH%"
nssm.exe set %SERVICE_NAME% AppDirectory "%INSTALL_DIR%"
nssm.exe set %SERVICE_NAME% Start SERVICE_AUTO_START
nssm.exe set %SERVICE_NAME% AppStdout "%INSTALL_DIR%\data\service.log"
nssm.exe set %SERVICE_NAME% AppStderr "%INSTALL_DIR%\data\service.log"

echo.
echo Service "%SERVICE_NAME%" installed. Start it with:
echo     nssm.exe start %SERVICE_NAME%
echo.
echo IMPORTANT: the first time this service starts, it generates a random
echo admin passphrase and writes it to data\service.log. Open that file to
echo retrieve it, then change it from the /admin settings page.

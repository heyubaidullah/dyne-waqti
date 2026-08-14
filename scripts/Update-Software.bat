@echo off
REM Pulls the latest code and restarts the service. Never touches data\ —
REM that directory is gitignored and lives next to the executable, not in
REM the git working tree, so a pull can never orphan or overwrite it.

set SERVICE_NAME=WaqtiService
set INSTALL_DIR=%~dp0..

cd /d "%INSTALL_DIR%"

echo Pulling latest changes...
git pull
if errorlevel 1 (
    echo ERROR: git pull failed. Aborting update ^(service left running^).
    exit /b 1
)

echo Rebuilding...
call make build-windows
if errorlevel 1 (
    echo ERROR: build failed. Aborting update ^(service left running on old binary^).
    exit /b 1
)

echo Restarting service...
nssm.exe restart %SERVICE_NAME%

echo Update complete. data\ was not touched.

@echo off
rem Windows build of the development web image. Tilt runs this from the repo
rem root with EXPECTED_REF set to the image tag it expects.
if "%EXPECTED_REF%"=="" set "EXPECTED_REF=inventory-manager/web:dev"
docker build -f infra/development/docker/web.Dockerfile -t %EXPECTED_REF% .

@echo off
SET EXECUTABLE=./main

IF "%1"=="build" (
    %EXECUTABLE% build
) ELSE IF "%1"=="dev" (
    %EXECUTABLE% dev
) ELSE (
    ECHO Usage: go-forth [build|dev]
)

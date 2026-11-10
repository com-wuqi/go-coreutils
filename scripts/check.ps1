#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Runs the repository checks: gofmt, go vet, go build, and go test.

.PARAMETER Tool
    Only run tests for one tool package, for example: -Tool cat.

.PARAMETER SkipTests
    Skip go test. Useful while a tool is still unimplemented.

.EXAMPLE
    pwsh -NoProfile -File scripts/check.ps1 -SkipTests

.EXAMPLE
    pwsh -NoProfile -File scripts/check.ps1 -Tool echo
#>
[CmdletBinding()]
param(
    [string]$Tool,
    [switch]$SkipTests
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
$failed = $false

Write-Host ''
Write-Host '==> gofmt' -ForegroundColor Cyan
$unformatted = @(& gofmt -l .)
if ($LASTEXITCODE -ne 0) {
    Write-Host "gofmt failed with exit code $LASTEXITCODE" -ForegroundColor Red
    $failed = $true
} elseif ($unformatted.Count -gt 0) {
    Write-Host 'Files that need formatting:' -ForegroundColor Red
    $unformatted | ForEach-Object { Write-Host "  $_" }
    $failed = $true
} else {
    Write-Host 'ok' -ForegroundColor Green
}

Write-Host ''
Write-Host '==> go vet' -ForegroundColor Cyan
& go vet ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host 'go vet failed' -ForegroundColor Red
    $failed = $true
} else {
    Write-Host 'ok' -ForegroundColor Green
}

Write-Host ''
Write-Host '==> go build' -ForegroundColor Cyan
& go build ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host 'go build failed' -ForegroundColor Red
    $failed = $true
} else {
    Write-Host 'ok' -ForegroundColor Green
}

if (-not $SkipTests) {
    Write-Host ''
    if ($Tool) {
        $target = './internal/' + $Tool
        Write-Host "==> go test $target" -ForegroundColor Cyan
        & go test $target
    } else {
        Write-Host '==> go test ./...' -ForegroundColor Cyan
        & go test ./...
    }
    if ($LASTEXITCODE -ne 0) {
        Write-Host 'go test failed' -ForegroundColor Red
        $failed = $true
    } else {
        Write-Host 'ok' -ForegroundColor Green
    }
}

Pop-Location

Write-Host ''
if ($failed) {
    Write-Host 'CHECK FAILED' -ForegroundColor Red
    exit 1
}
Write-Host 'CHECK PASSED' -ForegroundColor Green
exit 0

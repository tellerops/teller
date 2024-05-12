#!/usr/bin/env pwsh
# Copyright 2018 the Deno authors. All rights reserved. MIT license.
# TODO(everyone): Keep this script simple and easily auditable.

$ErrorActionPreference = 'Stop'

$Project = "teller"
$Repo = "tellerops/teller"

if ($v) {
  $Version = "v${v}"
}
if ($args.Length -eq 1) {
  $Version = $args.Get(0)
}

$Install = $env:BP_INSTALL
$BinDir = if ($Install) {
  "$Install"
} else {
  "$Home\.$Project-bin"
}

$Zip = "$BinDir\$Project.zip"
$Exe = "$BinDir\$Project.exe"
$Target = 'x86_64-windows'

# GitHub requires TLS 1.2
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$Uri = if (!$Version) {
  "https://github.com/$Repo/releases/latest/download/${Project}-${Target}.zip"
} else {
  "https://github.com/$Repo/releases/download/${Version}/$Project-${Target}.zip"
}

if (!(Test-Path $BinDir)) {
  New-Item $BinDir -ItemType Directory | Out-Null
}

curl.exe -Lo $Zip $Uri

tar.exe xf $Zip -C $BinDir --strip-components 1 

Remove-Item $Zip

$User = [EnvironmentVariableTarget]::User
$Path = [Environment]::GetEnvironmentVariable('Path', $User)
if (!(";$Path;".ToLower() -like "*;$BinDir;*".ToLower())) {
  [Environment]::SetEnvironmentVariable('Path', "$Path;$BinDir", $User)
  $Env:Path += ";$BinDir"
}

Write-Output "$Project was installed successfully to $Exe"
Write-Output "Run with '--help' to get started"

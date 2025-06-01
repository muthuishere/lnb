$ErrorActionPreference = 'Stop'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"
$url64 = 'https://github.com/muthuishere/lnb/releases/download/v0.1.0/lnb-windows-amd64.zip'

$packageArgs = @{
  packageName   = $env:ChocolateyPackageName
  unzipLocation = $toolsDir
  url64bit      = $url64
  checksum64    = 'SHA256 hash will be added during release'
  checksumType64= 'sha256'
}

Install-ChocolateyZipPackage @packageArgs

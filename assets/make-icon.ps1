# Renders assets\icon.svg to assets\icon-256.png (256x256, transparent corners)
# with the headless Microsoft Edge that ships with Windows.
# Run it only after changing icon.svg; the PNG is committed.
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

$edge = "${env:ProgramFiles(x86)}\Microsoft\Edge\Application\msedge.exe"
if (-not (Test-Path $edge)) { throw "Microsoft Edge not found: $edge" }

$tmp = Join-Path $env:TEMP ("kit-icon-" + [guid]::NewGuid())
New-Item -ItemType Directory $tmp | Out-Null
try {
    $svg = Get-Content (Join-Path $PSScriptRoot 'icon.svg') -Raw -Encoding UTF8
    $html = Join-Path $tmp 'icon.html'
    $png = Join-Path $tmp 'icon.png'
    $page = "<!doctype html><html><head><style>html,body{margin:0;background:transparent}svg{display:block;width:256px;height:256px}</style></head><body>$svg</body></html>"
    [IO.File]::WriteAllText($html, $page, (New-Object Text.UTF8Encoding $false))
    $edgeArgs = @('--headless', '--disable-gpu', '--hide-scrollbars', '--default-background-color=00000000',
        '--window-size=256,256', "--screenshot=$png", "--user-data-dir=$tmp\profile", ([Uri]$html).AbsoluteUri)
    Start-Process $edge -ArgumentList $edgeArgs -Wait -WindowStyle Hidden
    if (-not (Test-Path $png)) { throw "Edge did not write $png" }
    Copy-Item $png (Join-Path $PSScriptRoot 'icon-256.png') -Force
    Write-Host 'Wrote assets\icon-256.png'
} finally {
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

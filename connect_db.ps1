# PowerShell Script to automatically establish and maintain the database SSH tunnel
Stop-Process -Name plink -Force -ErrorAction SilentlyContinue

$ServerIp = "177.155.199.119"
$PortForward = "3306:127.0.0.1:3306"
$User = "root"
$Password = "u6m042uFYi9y8EHdRU"

Clear-Host
Write-Host "=====================================================" -ForegroundColor Cyan
Write-Host "     SexDaily SSH Database Tunnel Auto-Connector     " -ForegroundColor Cyan
Write-Host "=====================================================" -ForegroundColor Cyan
Write-Host "Forwarding local port 3306 to VPS MariaDB..." -ForegroundColor Yellow
Write-Host "Keep this window open to maintain the database connection." -ForegroundColor Gray
Write-Host ""

while ($true) {
    Write-Host "[$(Get-Date -Format 'HH:mm:ss')] Starting SSH tunnel..." -ForegroundColor White
    
    # Run plink.exe and block until it exits
    # -ssh: SSH protocol
    # -N: Do not open a shell/command screen (just forward port)
    # -L: Port forwarding spec
    # -pw: Password
    & plink.exe -ssh -N -L $PortForward "${User}@${ServerIp}" -pw $Password

    Write-Host "[$(Get-Date -Format 'HH:mm:ss')] SSH tunnel aborted or disconnected. Retrying in 5 seconds..." -ForegroundColor Red
    Start-Sleep -Seconds 5
}

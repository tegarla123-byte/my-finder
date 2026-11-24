$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..\..\..")
Push-Location $repoRoot
try {
    Write-Host "==> [1/4] Running go vet..." -ForegroundColor Cyan
    go vet ./...
    if ($LASTEXITCODE -ne 0) { throw "go vet failed with exit code $LASTEXITCODE" }

    Write-Host "==> [2/4] Running go test..." -ForegroundColor Cyan
    go test -v ./...
    if ($LASTEXITCODE -ne 0) { throw "go test failed with exit code $LASTEXITCODE" }

    Write-Host "==> [3/4] Building release binary (my-finder.exe)..." -ForegroundColor Cyan
    go build -ldflags="-H=windowsgui -s -w" -o my-finder.exe .
    if ($LASTEXITCODE -ne 0) { throw "go build failed with exit code $LASTEXITCODE" }

    $fi = Get-Item .\my-finder.exe
    Write-Host ("    Binary size: {0:N2} MB ({1} bytes)" -f ($fi.Length / 1MB), $fi.Length)

    Write-Host "==> [4/4] Live GUI, RAM, and clean shutdown smoke test..." -ForegroundColor Cyan
    Add-Type -TypeDefinition @"
using System;
using System.Text;
using System.Runtime.InteropServices;

public class MyFinderProbe {
    public delegate bool EnumProc(IntPtr h, IntPtr l);
    [StructLayout(LayoutKind.Sequential)] public struct RECT { public int L, T, R, B; }
    [DllImport("user32.dll")] public static extern bool EnumWindows(EnumProc p, IntPtr l);
    [DllImport("user32.dll")] public static extern uint GetWindowThreadProcessId(IntPtr h, out uint pid);
    [DllImport("user32.dll", CharSet = CharSet.Unicode)] public static extern int GetWindowText(IntPtr h, StringBuilder sb, int max);
    [DllImport("user32.dll", CharSet = CharSet.Unicode)] public static extern int GetClassName(IntPtr h, StringBuilder sb, int max);
    [DllImport("user32.dll")] public static extern bool IsWindowVisible(IntPtr h);
    [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr h, out RECT r);
    [DllImport("user32.dll")] public static extern bool PostMessage(IntPtr h, uint msg, IntPtr w, IntPtr l);

    public static string FindAndCloseMainWindow(uint targetPid) {
        string found = null;
        EnumWindows((h, l) => {
            uint pid;
            GetWindowThreadProcessId(h, out pid);
            if (pid == targetPid) {
                var cls = new StringBuilder(256);
                GetClassName(h, cls, 256);
                if (cls.ToString() == "MyFinderMainWindow") {
                    var title = new StringBuilder(256);
                    GetWindowText(h, title, 256);
                    RECT r;
                    GetWindowRect(h, out r);
                    bool vis = IsWindowVisible(h);
                    found = string.Format("HWND={0} Visible={1} Title='{2}' Size={3}x{4}", h, vis, title, r.R - r.L, r.B - r.T);
                    PostMessage(h, 0x0010, IntPtr.Zero, IntPtr.Zero); // WM_CLOSE
                    return false;
                }
            }
            return true;
        }, IntPtr.Zero);
        return found;
    }
}
"@ -ErrorAction SilentlyContinue

    $proc = Start-Process .\my-finder.exe -PassThru
    try {
        $null = $proc.WaitForInputIdle(3000)
        Start-Sleep -Milliseconds 500
        $proc.Refresh()
        if ($proc.HasExited) {
            throw "my-finder.exe exited prematurely with code $($proc.ExitCode)"
        }

        $perf = Get-CimInstance Win32_PerfFormattedData_PerfProc_Process -Filter "IDProcess = $($proc.Id)" -ErrorAction SilentlyContinue
        $winInfo = [MyFinderProbe]::FindAndCloseMainWindow([uint32]$proc.Id)
        if (-not $winInfo) {
            throw "MyFinderMainWindow was not found for PID $($proc.Id)"
        }

        Write-Host "    Window: $winInfo"
        if ($perf -and $perf.WorkingSetPrivate) {
            Write-Host ("    RAM (Task Manager / Private WS): {0:N2} MB" -f ([double]$perf.WorkingSetPrivate / 1MB))
        }

        Start-Sleep -Milliseconds 500
        $proc.Refresh()
        if (-not $proc.HasExited -or $proc.ExitCode -ne 0) {
            throw "Clean WM_CLOSE shutdown failed (HasExited=$($proc.HasExited), ExitCode=$($proc.ExitCode))"
        }
        Write-Host "    Shutdown: clean exit (ExitCode=0)" -ForegroundColor Green
    } finally {
        if (-not $proc.HasExited) {
            $proc.Kill()
        }
    }

    Write-Host "==> All checks passed!" -ForegroundColor Green
} finally {
    Pop-Location
}

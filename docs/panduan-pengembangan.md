# Panduan Pengembangan (Development Guide)

Dokumen ini memandu langkah demi langkah mulai dari menyiapkan environment development, menjalankan dan melakukan debugging aplikasi, hingga verifikasi dan build rilis **My Finder**.

---

## 1. Prasyarat & Setup Environment

### Kebutuhan Sistem
- **OS:** Windows 10 atau Windows 11 (64-bit / `amd64`)
- **Go:** Versi **1.24** atau lebih baru

> **Keunggulan setup My Finder:** Proyek ini menggunakan **pure Go standard library** (zero third-party dependencies, zero CGO). Anda **tidak** perlu menginstal GCC/MinGW, MSVC, Node.js, WebView2 SDK, maupun menjalankan `go mod download`.

### Instalasi Go (Jika Belum Terpasang)
Unduh installer Windows dari https://go.dev/dl/ atau instal lewat `winget`:

```powershell
winget install GoLang.Go
```

Setelah selesai, buka terminal PowerShell baru lalu verifikasi instalasi:

```powershell
go version
go env GOOS GOARCH
```

Pastikan output menampilkan Go `1.24+` dengan `GOOS=windows` dan `GOARCH=amd64`.

### Rekomendasi Editor
- **VS Code** / **Cursor** / **Antigravity IDE** dengan ekstensi resmi **Go (`golang.go`)** untuk autocomplete (`gopls`), navigasi definisi, dan auto-format (`gofmt`).

---

## 2. Langkah Menjalankan Aplikasi

Masuk ke direktori proyek terlebih dahulu:

```powershell
cd d:\LEARN\my-finder
```

Tersedia 3 cara menjalankan aplikasi sesuai kebutuhan:

### Cara A — Iterasi Cepat (`go run .`)
Cocok saat mengembangkan fitur baru atau menguji perubahan kode secara cepat:

```powershell
go run .
```

- Jendela aplikasi akan terbuka bersamaan dengan terminal aktif, sehingga output `fmt.Println` / `log.Println` maupun *panic stack trace* langsung terlihat di terminal.
- **Catatan lokasi `shortcuts.json`:** Karena `go run .` menaruh executable sementara di direktori `%TEMP%`, aplikasi otomatis menyimpan dan membaca data di `%APPDATA%\MyFinder\shortcuts.json` agar data uji tidak hilang saat folder temp dibersihkan.

### Cara B — Build Debug Portable (`my-finder-debug.exe`)
Cocok jika Anda ingin menguji mode **portable** (`shortcuts.json` di folder proyek) sambil tetap melihat log di konsol:

```powershell
go build -o my-finder-debug.exe .
.\my-finder-debug.exe
```

### Cara C — Build Rilis Produksi (`my-finder.exe`)
Menghasilkan 1 file `.exe` portable tanpa jendela terminal/CMD dan ukuran binary minimal (~2,3 MB):

```powershell
go build -ldflags="-H=windowsgui -s -w" -o my-finder.exe .
.\my-finder.exe
```

---

## 3. Alur Kerja Pengembangan Harian

1. **Edit kode** sesuai modul yang relevan (lihat peta modul di [dokumentasi-teknis.md](./dokumentasi-teknis.md#2-struktur-file--tanggung-jawab-modul)).
2. **Format kode:**
   ```powershell
   gofmt -w .
   ```
3. **Jalankan analisis statis & unit test:**
   ```powershell
   go vet ./...
   go test -v ./...
   ```
4. **Jalankan skrip verifikasi lengkap** (tutup dulu `my-finder.exe` jika sedang berjalan):
   ```powershell
   powershell -ExecutionPolicy Bypass -File .agents/skills/my-finder-win32-dev/scripts/verify.ps1
   ```
   Skrip ini otomatis menjalankan:
   - `go vet ./...`
   - `go test -v ./...`
   - Build rilis `my-finder.exe` + cek ukuran binary
   - Smoke test live GUI (`MyFinderMainWindow` tampil normal), pengukuran RAM Private Working Set, dan verifikasi penutupan bersih (`ExitCode=0`).

---

## 4. Debugging & Verifikasi Kualitas

### Melihat Log & Panic
Binary rilis dibangun dengan `-H=windowsgui` sehingga `stdout`/`stderr` tidak ditampilkan. Saat perlu debugging dengan `fmt.Println` atau melihat detail *panic*, jalankan lewat `go run .` atau `my-finder-debug.exe`.

### Mengecek Kebocoran GDI Handle & Penggunaan RAM
Karena UI digambar langsung dengan Win32 GDI, pastikan perubahan UI tidak menyebabkan *GDI handle leak*:
1. Jalankan `my-finder.exe`.
2. Buka **Task Manager → tab Details**.
3. Klik kanan header kolom → **Select columns** → centang **GDI objects** dan **Memory (active private working set)**.
4. Lakukan resize jendela berulang kali, scroll naik-turun, hover kartu, serta buka-tutup dialog **Tambah/Edit Shortcut**.
5. Pastikan jumlah **GDI objects** kembali stabil (tidak terus menanjak) dan RAM tetap rendah (~3–8 MB).

### Menguji Tema Adaptif & HiDPI
- **Tema Dark / Light Live:** Saat aplikasi terbuka, ubah mode di *Windows Settings → Personalization → Colors → Choose your mode*. Title bar, latar belakang, kotak pencarian, dan kartu harus langsung menyesuaikan tanpa restart.
- **HiDPI Scaling:** Ubah *Scale* di *Windows Settings → System → Display* (misal 100% ↔ 125% ↔ 150%) atau pindahkan jendela ke monitor dengan DPI berbeda; teks dan ikon harus tetap tajam dan proporsional.

---

## 5. Aturan Penting Saat Mengubah Kode

- **Jaga zero dependencies & zero CGO:** Jangan menambahkan module eksternal ke `go.mod` kecuali benar-benar disepakati.
- **Patuhi `go vet` (`unsafeptr`):** Jangan mengonversi `uintptr` dari Win32 (`lParam`, `GlobalLock`) langsung ke `unsafe.Pointer`. Gunakan `procRtlMoveMemory.Call(...)` di `win32.go`.
- **Disiplin GDI:** Setiap `CreateSolidBrush`, `CreatePen`, `CreateCompatibleBitmap`, `CreateCompatibleDC`, `CreateRectRgn`, atau `CreateFontW` wajib dibersihkan dengan `DeleteObject`/`DeleteDC` setelah mengembalikan handle lama lewat `SelectObject`.
- **Jangan lakukan I/O di `WM_PAINT`:** Hindari `os.Stat` atau akses disk/jaringan di dalam loop render agar UI tidak pernah macet.
- **Perbarui test & dokumentasi:** Tambahkan unit test di `store_test.go` untuk logika data/path, dan perbarui dokumen di `docs/` bila ada perubahan fitur atau arsitektur.

---

## 6. Troubleshooting Umum

| Masalah | Penyebab | Solusi |
| --- | --- | --- |
| `open my-finder.exe: Access is denied.` saat `go build` | Aplikasi `my-finder.exe` masih terbuka sehingga file dikunci Windows. | Tutup jendela **My Finder** terlebih dahulu, lalu jalankan `go build` ulang. |
| `go vet` error: `possible misuse of unsafe.Pointer` | Cast langsung dari `uintptr` Win32 ke `unsafe.Pointer`. | Salin struct/buffer menggunakan `procRtlMoveMemory.Call(...)`. |
| Perubahan `shortcuts.json` di folder proyek tidak muncul saat `go run .` | `go run .` berjalan dari `%TEMP%` sehingga memakai fallback `%APPDATA%\MyFinder\shortcuts.json`. | Gunakan `go build -o my-finder-debug.exe .` lalu jalankan dari folder proyek untuk menguji mode portable. |
| PowerShell menolak menjalankan `verify.ps1` | Kebijakan *ExecutionPolicy* default Windows membatasi skrip lokal. | Jalankan dengan flag `powershell -ExecutionPolicy Bypass -File ...`. |

---

## 7. CI/CD & Merilis Versi Baru (GitHub Actions)

Proyek ini memiliki dua workflow otomatis di `.github/workflows/`:

| Workflow | Pemicu (*Trigger*) | Yang Dilakukan |
| --- | --- | --- |
| **CI** (`.github/workflows/ci.yml`) | Setiap `git push` atau Pull Request ke branch `main` | Menjalankan `go vet`, `go test`, dan verifikasi build `my-finder.exe` (**tidak** membuat halaman Release). |
| **Auto-Release** (`.github/workflows/release.yml`) | Hanya saat push **tag** berawalan `v` (misal `v1.0.0`) | Menjalankan `go vet` & `go test`, mem-build `my-finder.exe` (`-trimpath -ldflags="-H=windowsgui -s -w"`), membuat `checksums.txt` (SHA256), lalu menerbitkan halaman **GitHub Release** lengkap dengan catatan rilis otomatis. |

> Selama Anda hanya melakukan `git commit` dan `git push origin main` untuk pengembangan sehari-hari, halaman Release **tidak** akan dibuat.

### Langkah Menerbitkan Rilis Baru
Setiap kali kode sudah siap dirilis ke pengguna, buat tag versi baru (harus berbeda dari versi sebelumnya) lalu push ke GitHub:

```powershell
git tag v1.0.0
git push origin main
git push origin v1.0.0
```

Untuk rilis-rilis berikutnya, gunakan pola **Semantic Versioning (`vMAJOR.MINOR.PATCH`)**:
- `v1.0.1` — perbaikan bug kecil (*patch*)
- `v1.1.0` — penambahan fitur baru (*minor*)
- `v2.0.0` — perubahan besar / tidak kompatibel (*major*)

Catatan rilis (*release notes*) akan otomatis dibuat oleh GitHub dari daftar commit/PR sejak tag versi sebelumnya.

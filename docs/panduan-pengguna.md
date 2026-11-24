# Panduan Pengguna My Finder (Non-Teknis)

**My Finder** adalah aplikasi desktop Windows berukuran sangat kecil dan hemat RAM yang membantu Anda menyimpan daftar lokasi folder (atau file) penting, lalu langsung membukanya di **jendela File Explorer baru** cukup dengan **sekali klik**.

---

## 1. Mengapa Menggunakan My Finder?

- **Sekali klik langsung sampai tujuan:** Tidak perlu bolak-balik membuka File Explorer dan menelusuri folder bertingkat.
- **Membuka jendela baru:** Setiap kali item diklik, My Finder membuka jendela File Explorer baru langsung di dalam folder yang Anda daftarkan.
- **Super ringan & ramah RAM:** Hanya memakai sekitar **4–5 MB RAM** dan berukuran sekitar **2,3 MB** (1 file saja), tanpa perlu instalasi rumit.
- **Portable:** Cukup salin `my-finder.exe` (dan file `shortcuts.json` jika sudah ada) ke folder atau flashdisk mana pun.
- **Tampilan nyaman di mata:** Otomatis mengikuti tema **Dark Mode** atau **Light Mode** Windows Anda.

---

## 2. Cara Memulai

1. Buka folder tempat `my-finder.exe` berada.
2. Klik ganda **`my-finder.exe`**.
3. Jendela **My Finder** akan langsung terbuka (tanpa jendela terminal/CMD).

> **Tips:** Anda bisa klik kanan `my-finder.exe` lalu pilih **Pin to taskbar** atau **Pin to Start** agar bisa diakses kapan saja dengan 1 klik dari taskbar Windows.

---

## 3. Cara Menambah Shortcut

Ada dua cara cepat untuk mendaftarkan folder ke My Finder:

### Cara A — Tombol `+ Tambah`
1. Klik tombol **`+ Tambah`** di pojok kanan atas (atau tekan `Ctrl + N`).
2. Pada kolom **Lokasi Folder**, klik tombol **`Pilih Folder...`** untuk memilih folder lewat jendela Windows, atau tempelkan (paste) alamat folder secara langsung.
3. Isi kolom **Nama Shortcut** sesuai keinginan Anda (misalnya: *Proyek Kantor*, *Dokumen Pajak*, *Downloads*).
   - Jika dikosongkan, My Finder akan otomatis memakai nama folder tersebut.
4. Klik **`Simpan`** (atau tekan `Enter`).

### Cara B — Drag & Drop (Tarik & Lepas)
1. Buka File Explorer dan pilih satu atau beberapa folder sekaligus.
2. Tarik (drag) folder tersebut dan lepaskan langsung ke jendela **My Finder**.
3. Folder otomatis masuk ke daftar shortcut Anda (folder yang sudah pernah terdaftar tidak akan diduplikasi).

---

## 4. Cara Membuka Shortcut

- **Pakai Mouse:** Cukup **klik 1 kali (single-click)** pada kartu shortcut yang diinginkan. Jendela File Explorer baru akan langsung terbuka di lokasi folder tersebut.
- **Pakai Keyboard:** Gunakan tombol panah `↑` / `↓` untuk memilih item, lalu tekan **`Enter`**.

> **Catatan untuk Path File:** Selain folder, jika Anda mendaftarkan path menuju sebuah **file**, mengkliknya akan membuka jendela File Explorer baru dan langsung menyorot (select) file tersebut.

> **Tips Variabel Windows:** Anda juga bisa menuliskan path dinamis seperti `%USERPROFILE%\Downloads` atau `%APPDATA%`. My Finder akan otomatis menerjemahkannya saat dibuka.

---

## 5. Pencarian / Filter Cepat

Jika daftar shortcut Anda sudah banyak:
1. Langsung ketik kata kunci di kotak **Cari shortcut...** di bagian atas (atau tekan `Ctrl + F` untuk melompat ke kotak pencarian).
2. Pencarian mencocokkan **nama shortcut** maupun **alamat folder (path)**.
3. Anda bisa mengetik lebih dari satu kata sekaligus (misal: `learn finder` untuk menemukan `D:\LEARN\my-finder`).
4. Tekan **`Enter`** untuk langsung membuka hasil yang terpilih, atau tekan **`Esc`** / klik tombol **`×`** untuk menghapus pencarian.

---

## 6. Mengedit, Mengatur Urutan, dan Menghapus Shortcut

### Tombol Cepat di Kartu (Saat Kursor Diarahkan / Hover)
Saat Anda mengarahkan mouse ke salah satu item, dua ikon kecil akan muncul di sisi kanan:
- **Ikon Pensil** — Membuka jendela untuk mengubah nama atau lokasi folder.
- **Ikon Silang (`×`)** — Menghapus shortcut (akan muncul konfirmasi terlebih dahulu agar tidak terhapus tidak sengaja).

### Menu Klik Kanan
Klik kanan pada item mana saja untuk membuka menu lengkap:
- **Buka di Explorer Baru** — Membuka folder di jendela File Explorer baru.
- **Salin Path** — Menyalin alamat folder ke clipboard (siap di-paste ke terminal/aplikasi lain).
- **Edit Shortcut...** — Mengubah nama atau lokasi path.
- **Geser ke Atas / Geser ke Bawah** — Mengubah posisi urutan item (tersedia saat kotak pencarian kosong).
- **Hapus** — Menghapus item dari daftar.

---

## 7. Daftar Pintasan Keyboard

| Pintasan | Fungsi |
| --- | --- |
| `Ctrl + F` | Fokus ke kotak pencarian dan memilih seluruh teks |
| `Ctrl + N` | Membuka dialog **Tambah Shortcut Baru** |
| `↑` / `↓` | Berpindah pilihan item ke atas / ke bawah |
| `Enter` | Membuka item terpilih di jendela File Explorer baru |
| `Alt + ↑` / `Alt + ↓` | Memindahkan urutan item terpilih ke atas / ke bawah |
| `Esc` | Menghapus teks di kotak pencarian |
| `Delete` | Menghapus item terpilih (saat fokus tidak di kotak ketik) |

*Di dalam dialog Tambah / Edit Shortcut:*
- `Tab` berpindah antara kolom Nama dan Lokasi Folder.
- `Enter` menyimpan perubahan.
- `Esc` membatalkan dan menutup dialog.

---

## 8. Penyimpanan Data & Cara Backup

- Secara default, seluruh daftar shortcut Anda disimpan otomatis ke file **`shortcuts.json`** di folder yang sama dengan `my-finder.exe`.
- Jika `my-finder.exe` diletakkan di folder yang tidak bisa ditulisi (misalnya `C:\Program Files`), data otomatis disimpan di `%APPDATA%\MyFinder\shortcuts.json`.
- **Untuk backup atau pindah perangkat:** Cukup salin file `shortcuts.json` bersama `my-finder.exe`.

---

## 9. Tanya Jawab (FAQ)

**Q: Muncul pesan "lokasi tidak ditemukan" saat item diklik. Kenapa?**
A: Folder yang didaftarkan mungkin sudah dipindahkan, diganti namanya, atau berada di drive eksternal (USB/harddisk) yang sedang tidak terpasang. Anda bisa mengedit path item tersebut lewat tombol **Edit**.

**Q: Apakah menghapus shortcut di My Finder ikut menghapus folder asli saya?**
A: **Tidak.** Menghapus item di My Finder hanya menghapus catatan daftarnya di `shortcuts.json`; folder dan file asli Anda di komputer tetap aman sepenuhnya.

**Q: Bagaimana cara mengganti tampilan terang / gelap?**
A: My Finder otomatis mengikuti pengaturan **Settings → Personalization → Colors → Choose your mode** di Windows. Saat Anda mengganti tema Windows, tampilan My Finder akan langsung menyesuaikan tanpa perlu restart aplikasi.

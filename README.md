# PEDE Golang Agentic AI

Modul ini adalah antar-muka (*interface*) utama tempat di mana **AI Agent** berinteraksi dengan pengguna (Anda). Agen Golang ini dilengkapi dengan fitur cerdas (Function Calling) yang memungkinkannya berpikir dan secara otonom memanggil *Database* Python (Qdrant) saat ia butuh mencari data akademik.

## Prasyarat
1. Pastikan Anda telah mengatur *environment variable* untuk kunci API Gemini:
   ```bash
   # Windows PowerShell
   $env:GEMINI_API_KEY="AIzaSy..."
   ```
2. Pastikan Server Python FastAPI sudah berjalan di *background*. Buka terminal lain dan jalankan perintah ini di folder `pede`:
   ```bash
   uvicorn api:app --port 8000
   ```

## Cara Menjalankan Agen
Masuk ke dalam folder `agent-go` ini, lalu jalankan:
```bash
go run .
```

## Rencana Verifikasi (Verification Plan)
Untuk memastikan *Agentic AI* ini berfungsi sempurna, lakukan 2 pengujian berikut di terminal *chatbot* Anda:

**Tes 1: Percakapan Biasa (Non-RAG)**
- **Anda:** "Halo, siapa kamu?"
- **Ekspektasi:** Agen harus bisa menjawab dengan ramah (misal: "Halo, saya adalah asisten akademik...") **TANPA** memanggil fungsi pencarian database, karena pertanyaannya tidak bersifat akademis.

**Tes 2: Percakapan Akademik (Agentic RAG)**
- **Anda:** "Tolong cari di database, apa hasil eksperimen dari neurosymbolic?"
- **Ekspektasi:** 
  1. Anda akan melihat log `[🤖 AGEN BERPIKIR] Memanggil fungsi pencarian RAG...`
  2. Golang akan mengirim HTTP Request ke `localhost:8000`.
  3. Setelah teks diterima, AI akan merangkum teks tersebut dan memberikannya kepada Anda dalam Bahasa Indonesia yang baik dan profesional.

*(Catatan: Anda juga bisa memberikan DOI spesifik di pertanyaan Anda, misalnya: "Carikan kesimpulan dari jurnal ber-DOI 10.1016/j.inpa.2026.02.006" dan agen akan otomatis memfilter pencarian tersebut!)*

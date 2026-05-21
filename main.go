package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Define the tool function structure exactly as Gemini expects
func queryScientificDatabaseFunc(query string, doi string) string {
	fmt.Printf("\n[🤖 AGEN BERPIKIR] Memanggil fungsi pencarian RAG...\n")
	fmt.Printf("   -> Kueri Pencarian : %s\n", query)
	if doi != "" {
		fmt.Printf("   -> Target DOI      : %s\n", doi)
	}

	result, err := SearchDatabase(query, doi, 5)
	if err != nil {
		fmt.Printf("   -> ERROR: %v\n", err)
		return fmt.Sprintf("Error saat mencari di database: %v", err)
	}
	
	fmt.Printf("   -> Sukses! Teks hasil dari database telah diterima dan sedang dianalisis...\n")
	return result
}

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ ERROR: GEMINI_API_KEY tidak ditemukan di environment variables.")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Gagal membuat klien Gemini: %v", err)
	}
	defer client.Close()

	// 1. Setup function declaration untuk Agentic Tool
	searchFunction := &genai.FunctionDeclaration{
		Name:        "query_scientific_database",
		Description: "Gunakan alat ini SELALU saat pengguna bertanya tentang informasi dari jurnal, artikel ilmiah, atau mencari data akademik tertentu. Alat ini akan mencari jawaban di dalam database jurnal lokal secara otomatis.",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"query": {
					Type:        genai.TypeString,
					Description: "Pertanyaan atau kata kunci yang dicari. Harus diubah ke dalam bahasa Inggris yang baik untuk hasil pencarian optimal.",
				},
				"doi": {
					Type:        genai.TypeString,
					Description: "Optional. Digital Object Identifier (DOI) dari artikel ilmiah tertentu jika pengguna menyebutkannya (contoh: 10.1016/j.inpa.2026.02.006). Biarkan kosong jika tidak disebutkan.",
				},
			},
			Required: []string{"query"},
		},
	}

	// 2. Setup Model
	model := client.GenerativeModel("gemini-2.5-flash")
	model.Tools = []*genai.Tool{
		{FunctionDeclarations: []*genai.FunctionDeclaration{searchFunction}},
	}
	
	// 3. System Instruction
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text("Anda adalah asisten akademik cerdas berbasis Golang. Anda HANYA menjawab pertanyaan berdasarkan data ilmiah. Jika pengguna bertanya hal akademis, ANDA WAJIB memanggil fungsi query_scientific_database. Jika fungsi tersebut mengembalikan data, gunakan data tersebut untuk merangkai jawaban akhir (RAG). Jangan pernah mengarang data (halusinasi). Jawab dengan ramah, profesional, dan dalam bahasa Indonesia."),
		},
	}

	// Mulai sesi chat yang mengingat konteks
	session := model.StartChat()

	fmt.Println("=========================================================")
	fmt.Println("🚀 PEDE Golang Agentic AI (Gemini 2.5) Siap!")
	fmt.Println("Ketik 'exit' atau 'quit' untuk keluar.")
	fmt.Println("Pastikan server Python (uvicorn api:app --port 8000) sudah berjalan!")
	fmt.Println("=========================================================")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\nAnda: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		if strings.ToLower(input) == "exit" || strings.ToLower(input) == "quit" {
			fmt.Println("Sampai jumpa!")
			break
		}

		// Kirim pesan awal ke Gemini
		resp, err := session.SendMessage(ctx, genai.Text(input))
		if err != nil {
			fmt.Printf("❌ Error menghubungi Gemini: %v\n", err)
			continue
		}

		// Agentic Loop: Handle response or Function Calls
		for {
			var functionCalls []genai.FunctionCall
			var responseTexts []string

			if resp.Candidates != nil && len(resp.Candidates) > 0 {
				for _, part := range resp.Candidates[0].Content.Parts {
					if call, ok := part.(genai.FunctionCall); ok {
						functionCalls = append(functionCalls, call)
					}
					if text, ok := part.(genai.Text); ok {
						responseTexts = append(responseTexts, string(text))
					}
				}
			}

			// Jika tidak ada panggilan fungsi, berarti agen ingin berbicara biasa ke pengguna
			if len(functionCalls) == 0 {
				if len(responseTexts) > 0 {
					fmt.Printf("\n🤖 Agen:\n%s\n", strings.Join(responseTexts, "\n"))
				}
				break // Keluar dari loop fungsi
			}

			// Jika agen meminta mengeksekusi alat (Tool)
			for _, call := range functionCalls {
				if call.Name == "query_scientific_database" {
					var query string
					var doi string
					
					if q, ok := call.Args["query"].(string); ok {
						query = q
					}
					if d, ok := call.Args["doi"].(string); ok {
						doi = d
					}

					// 1. Eksekusi fungsi Golang kita!
					functionResultString := queryScientificDatabaseFunc(query, doi)

					// 2. Format balikan agar sesuai standar JSON untuk Gemini
					resultMap := map[string]interface{}{
						"search_results": functionResultString,
					}
					
					// 3. Kirim kembali hasil database tersebut ke Gemini
					resp, err = session.SendMessage(ctx, genai.FunctionResponse{
						Name: call.Name,
						Response: resultMap,
					})
					
					if err != nil {
						fmt.Printf("❌ Error mengirim hasil fungsi kembali ke Gemini: %v\n", err)
						break
					}
				}
			}
		}
	}
}

<div class="bg-[var(--background)] border border-[var(--border)] rounded-lg overflow-hidden transition-all max-w-md"><div class="flex items-center gap-3 p-3"><div class="w-10 h-10 rounded-lg overflow-hidden flex-shrink-0 bg-[var(--accent)] flex items-center justify-center"><svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-plus text-white"><path d="M5 12h14"></path><path d="M12 5v14"></path></svg></div><div class="flex-1 min-w-0"><div class="flex items-center gap-2"><h3 class="text-sm font-semibold text-[var(--text)] truncate">ollama</h3><span class="flex items-center gap-1 text-xs text-green-500">✓ Configured</span></div><p class="text-xs text-[var(--text-muted)] mt-0.5 truncate">http://127.0.0.1:11434</p></div><div class="flex items-center gap-1.5 flex-shrink-0"><button class="p-1.5 hover:bg-[var(--surface)] rounded-md transition-colors text-[var(--text)]" title="Collapse"><svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-chevron-up"><path d="m18 15-6-6-6 6"></path></svg></button><button class="p-1.5 hover:bg-[var(--surface)] rounded-md transition-colors text-[var(--text)]" title="Edit"><svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-square-pen"><path d="M12 3H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.375 2.625a1 1 0 0 1 3 3l-9.013 9.014a2 2 0 0 1-.853.505l-2.873.84a.5.5 0 0 1-.62-.62l.84-2.873a2 2 0 0 1 .506-.852z"></path></svg></button><button class="p-1.5 hover:bg-red-500/20 rounded-md transition-colors text-red-500" title="Delete"><svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-trash2 lucide-trash-2"><path d="M3 6h18"></path><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path><line x1="10" x2="10" y1="11" y2="17"></line><line x1="14" x2="14" y1="11" y2="17"></line></svg></button></div></div><div class="border-t border-[var(--border)] p-4 bg-[var(--surface)] space-y-4"><div><label class="block text-sm font-medium text-[var(--text)] mb-2">Name</label><input type="text" class="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors" placeholder="My AI Provider" value="ollama"></div><div><label class="block text-sm font-medium text-[var(--text)] mb-2">Base URL</label><input type="url" class="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors" placeholder="providers.baseUrlPlaceholder" value="http://127.0.0.1:11434"></div><div><label class="flex items-center gap-3 cursor-pointer"><input type="checkbox" class="w-4 h-4 text-[var(--accent)] bg-[var(--background)] border-[var(--border)] rounded focus:ring-[var(--accent)] focus:ring-2" checked=""><div><div class="text-sm font-medium text-[var(--text)]">API Key not required</div><div class="text-xs text-[var(--text-muted)]">Disable if provider doesn't require API key (e.g., local Ollama)</div></div></label></div><div><label class="block text-sm font-medium text-[var(--text)] mb-2">API Key</label><input type="password" class="w-full px-3 py-2 bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors" placeholder="sk-..." value="ollama"></div><div class="text-xs text-[var(--text-muted)] bg-[var(--background)] border border-[var(--border)] rounded-md p-2">Debug: name="ollama" baseUrl="http://127.0.0.1:11434" requiresApiKey= apiKey="ollama..."</div><div class="flex items-center gap-2 pt-2"><button class="px-4 py-2 bg-[var(--accent)] hover:bg-[var(--accent)]/90 disabled:bg-gray-500 disabled:cursor-not-allowed text-white text-sm font-medium rounded-md transition-colors">Save</button><button class="px-4 py-2 text-sm font-medium text-[var(--text)] hover:bg-[var(--background)] rounded-md transition-colors">Cancel</button></div></div></div>


Такой сейчас html я хочу в него добавить такое:


 --------------------------
|  auto (openai)  |  curl  |
 --------------------------


 При выборе curl появляется такое 


 --------------------------
|  auto (openai)  |  curl  |
 --------------------------


 -------------------------------------------------
| 1 | curl http://<AI.EXEMPLE> \
| 1 |  -H "Content-Type: application/json" \
| 1 |  -H "Authorization: Bearer $APIKEY" \
| 1 |  -d '{
| 1 |    "model": $MODEL,
| 1 |    "messages": [{"role": "user", "content": 
| 1 |       "Hay !"}],
| 1 |    "stream": false
| 1 |  }'
 -------------------------------------------------
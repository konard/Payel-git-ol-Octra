import { useState } from 'react';
import { Send } from 'lucide-react';

interface ChatInputProps {
  onSendMessage: (message: string) => void;
}

export function ChatInput({ onSendMessage }: ChatInputProps) {
  const [inputValue, setInputValue] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (inputValue.trim()) {
      onSendMessage(inputValue.trim());
      setInputValue('');
    }
  };

  return (
    <div className="bg-[var(--surface)] border-t border-[var(--border)] p-4">
      <form onSubmit={handleSubmit} className="flex items-end gap-2">
        <div className="flex-1">
          <textarea
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            rows={2}
            className="w-full px-4 py-3 bg-[var(--background)] border border-[var(--border)] rounded-xl text-[var(--text)] text-sm resize-none placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
            placeholder="Введите сообщение..."
          />
        </div>
        <button
          type="submit"
          disabled={!inputValue.trim()}
          className={`flex items-center justify-center w-12 h-12 rounded-xl transition-colors flex-shrink-0 ${
            inputValue.trim()
              ? 'bg-[var(--accent)] hover:bg-[var(--accent-hover)] text-white'
              : 'bg-[var(--surface)] text-[var(--text-muted)] cursor-not-allowed'
          }`}
        >
          <Send size={20} />
        </button>
      </form>
    </div>
  );
}
import { useState, useRef, useEffect } from 'react';
import type { KeyboardEvent } from 'react';
import Markdown from 'react-markdown';

interface Message {
  role: 'user' | 'assistant';
  content: string;
}

const SUGGESTIONS = [
  "What's my current portfolio breakdown?",
  "Which position has the highest concentration risk?",
  "Suggest how I should rebalance my holdings",
  "What happened since my last sync?",
];

export default function Chat() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [streaming, setStreaming] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);
  const conversationId = useRef('conv-' + Date.now());

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const sendMessage = async (text: string) => {
    if (!text.trim() || streaming) return;

    const userMessage = text.trim();
    setInput('');
    setMessages(prev => [...prev, { role: 'user', content: userMessage }]);
    setStreaming(true);
    setMessages(prev => [...prev, { role: 'assistant', content: '' }]);

    try {
      const res = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message: userMessage,
          conversation_id: conversationId.current,
        }),
      });

      const reader = res.body?.getReader();
      const decoder = new TextDecoder();
      if (!reader) throw new Error('No reader');

      let buffer = '';
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6);
            if (data === '[DONE]') continue;
            try {
              const parsed = JSON.parse(data);
              if (parsed.text) {
                setMessages(prev => {
                  const updated = [...prev];
                  const last = updated[updated.length - 1];
                  if (last.role === 'assistant') {
                    updated[updated.length - 1] = { ...last, content: last.content + parsed.text };
                  }
                  return updated;
                });
              }
            } catch { /* skip */ }
          }
        }
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Unknown error';
      setMessages(prev => {
        const updated = [...prev];
        const last = updated[updated.length - 1];
        if (last.role === 'assistant' && !last.content) {
          updated[updated.length - 1] = { ...last, content: `Error: ${message}` };
        }
        return updated;
      });
    } finally {
      setStreaming(false);
    }
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage(input);
    }
  };

  return (
    <div className="chat-container">
      <div className="page-header animate-in" style={{ marginBottom: '1rem' }}>
        <div>
          <h1 className="page-title">AI Chat</h1>
          <p className="page-subtitle">Discuss your portfolio with Claude</p>
        </div>
      </div>

      <div className="chat-messages animate-in animate-in-1">
        {messages.length === 0 && (
          <div className="chat-welcome">
            <div style={{ fontSize: '2.5rem', marginBottom: '0.5rem', opacity: 0.3 }}>&#x2B22;</div>
            <h3>Portfolio Intelligence</h3>
            <p style={{ fontSize: '0.85rem', maxWidth: '400px', margin: '0 auto' }}>
              Ask anything about your holdings, risk exposure, rebalancing strategies, or market analysis.
            </p>
            <div className="chat-suggestions">
              {SUGGESTIONS.map((s, i) => (
                <button key={i} className="chat-suggestion" onClick={() => sendMessage(s)}>
                  {s}
                </button>
              ))}
            </div>
          </div>
        )}
        {messages.map((msg, i) => (
          <div key={i} className={`chat-bubble chat-bubble-${msg.role}`}>
            {msg.role === 'assistant' ? (
              <Markdown>{msg.content || (streaming && i === messages.length - 1 ? '...' : '')}</Markdown>
            ) : (
              <div style={{ whiteSpace: 'pre-wrap' }}>{msg.content}</div>
            )}
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      <div className="chat-input-row animate-in animate-in-2">
        <textarea
          className="chat-textarea"
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Ask about your portfolio..."
          disabled={streaming}
          rows={2}
        />
        <button
          className="btn btn-primary"
          onClick={() => sendMessage(input)}
          disabled={streaming || !input.trim()}
          style={{ height: '100%', minHeight: '56px' }}
        >
          {streaming ? <span className="spinner" /> : <SendIcon />}
        </button>
      </div>
    </div>
  );
}

function SendIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <line x1="22" y1="2" x2="11" y2="13" />
      <polygon points="22 2 15 22 11 13 2 9 22 2" />
    </svg>
  );
}

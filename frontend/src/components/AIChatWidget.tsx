import { useState, useRef, useEffect } from 'react';
import type { KeyboardEvent } from 'react';
import Markdown from 'react-markdown';

interface Message {
  role: 'user' | 'assistant';
  content: string;
}

const SUGGESTIONS = [
  "What's my portfolio breakdown?",
  "Which position has the highest risk?",
  "How should I rebalance?",
  "What changed since last sync?",
];

export default function AIChatWidget() {
  const [open, setOpen] = useState(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState('');
  const [streaming, setStreaming] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const conversationId = useRef('conv-' + Date.now());

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  useEffect(() => {
    if (open && inputRef.current) {
      inputRef.current.focus();
    }
  }, [open]);

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

  const clearChat = () => {
    setMessages([]);
    conversationId.current = 'conv-' + Date.now();
  };

  return (
    <>
      <button
        className={`ai-fab ${open ? 'ai-fab-hidden' : ''}`}
        onClick={() => setOpen(true)}
        aria-label="Open AI Chat"
      >
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M12 2L2 7l10 5 10-5-10-5z" />
          <path d="M2 17l10 5 10-5" />
          <path d="M2 12l10 5 10-5" />
        </svg>
        <span className="ai-fab-label">Ask AI</span>
      </button>

      <div className={`ai-panel ${open ? 'ai-panel-open' : ''}`}>
        <div className="ai-panel-header">
          <div className="ai-panel-header-left">
            <div className="ai-panel-avatar">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 2L2 7l10 5 10-5-10-5z" />
                <path d="M2 17l10 5 10-5" />
                <path d="M2 12l10 5 10-5" />
              </svg>
            </div>
            <div>
              <div className="ai-panel-title">FlowPilot AI</div>
              <div className="ai-panel-subtitle">Portfolio Intelligence</div>
            </div>
          </div>
          <div className="ai-panel-header-actions">
            {messages.length > 0 && (
              <button className="ai-panel-btn" onClick={clearChat} title="New conversation">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M21 2v6h-6" /><path d="M3 12a9 9 0 0 1 15-6.7L21 8" />
                  <path d="M3 22v-6h6" /><path d="M21 12a9 9 0 0 1-15 6.7L3 16" />
                </svg>
              </button>
            )}
            <button className="ai-panel-btn" onClick={() => setOpen(false)} title="Close">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M18 6L6 18" /><path d="M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <div className="ai-panel-messages">
          {messages.length === 0 && (
            <div className="ai-panel-welcome">
              <div className="ai-panel-welcome-icon">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <path d="M12 2L2 7l10 5 10-5-10-5z" />
                  <path d="M2 17l10 5 10-5" />
                  <path d="M2 12l10 5 10-5" />
                </svg>
              </div>
              <h4>Ask anything about your portfolio</h4>
              <p>I can analyze your holdings, suggest rebalancing strategies, and provide insights on your financial data.</p>
              <div className="ai-panel-suggestions">
                {SUGGESTIONS.map((s, i) => (
                  <button key={i} className="ai-panel-suggestion" onClick={() => sendMessage(s)}>
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <polyline points="9 18 15 12 9 6" />
                    </svg>
                    {s}
                  </button>
                ))}
              </div>
            </div>
          )}
          {messages.map((msg, i) => (
            <div key={i} className={`ai-msg ai-msg-${msg.role}`}>
              {msg.role === 'assistant' && (
                <div className="ai-msg-avatar">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M12 2L2 7l10 5 10-5-10-5z" />
                    <path d="M2 17l10 5 10-5" />
                    <path d="M2 12l10 5 10-5" />
                  </svg>
                </div>
              )}
              <div className={`ai-msg-bubble ai-msg-bubble-${msg.role}`}>
                {msg.role === 'assistant' ? (
                  <Markdown>{msg.content || (streaming && i === messages.length - 1 ? '...' : '')}</Markdown>
                ) : (
                  <div style={{ whiteSpace: 'pre-wrap' }}>{msg.content}</div>
                )}
              </div>
            </div>
          ))}
          <div ref={bottomRef} />
        </div>

        <div className="ai-panel-input">
          <div className="ai-panel-input-row">
            <textarea
              ref={inputRef}
              className="ai-panel-textarea"
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Ask about your portfolio..."
              disabled={streaming}
              rows={1}
            />
            <button
              className="ai-panel-send"
              onClick={() => sendMessage(input)}
              disabled={streaming || !input.trim()}
            >
              {streaming ? (
                <span className="spinner" style={{ width: 14, height: 14 }} />
              ) : (
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="22" y1="2" x2="11" y2="13" />
                  <polygon points="22 2 15 22 11 13 2 9 22 2" />
                </svg>
              )}
            </button>
          </div>
          <div className="ai-panel-footer">
            Powered by Claude &middot; FlowPilot AI
          </div>
        </div>
      </div>

      {open && <div className="ai-backdrop" onClick={() => setOpen(false)} />}
    </>
  );
}

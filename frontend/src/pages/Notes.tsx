import { useState, useEffect, useCallback } from 'react';
import { api } from '../api/client';
import type { Note } from '../types';
import NoteEditor from '../components/NoteEditor';

export default function Notes() {
  const [notes, setNotes] = useState<Note[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [filterTag, setFilterTag] = useState<string | null>(null);
  const [error, setError] = useState('');
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);

  const fetchNotes = useCallback(async () => {
    try {
      const data = await api.getNotes();
      setNotes(data.notes || []);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    }
  }, []);

  useEffect(() => { fetchNotes(); }, [fetchNotes]);

  const selectedNote = notes.find(n => n.id === selectedId) ?? null;

  const allTags = Array.from(new Set(notes.flatMap(n => n.tags || [])));

  const filteredNotes = filterTag
    ? notes.filter(n => (n.tags || []).includes(filterTag))
    : notes;

  const grouped = new Map<string, Note[]>();
  for (const note of filteredNotes) {
    const key = note.symbol || 'General';
    const arr = grouped.get(key) ?? [];
    arr.push(note);
    grouped.set(key, arr);
  }
  const sortedGroups = Array.from(grouped.entries()).sort((a, b) => {
    if (a[0] === 'General') return -1;
    if (b[0] === 'General') return 1;
    return a[0].localeCompare(b[0]);
  });

  const handleNewNote = async () => {
    try {
      const data = await api.createNote({
        title: 'Untitled Note',
        content: '',
        symbol: '',
        tags: [],
        pinned: false,
      });
      await fetchNotes();
      setSelectedId(data.note.id);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    }
  };

  const handleUpdateField = async (field: string, value: unknown) => {
    if (!selectedNote) return;
    try {
      const data = await api.updateNote(selectedNote.id, { ...selectedNote, [field]: value });
      setNotes(prev => prev.map(n => n.id === data.note.id ? data.note : n));
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    }
  };

  const handleContentUpdate = useCallback((content: string) => {
    if (!selectedId) return;
    const current = notes.find(n => n.id === selectedId);
    if (!current) return;
    api.updateNote(selectedId, { ...current, content }).then(data => {
      setNotes(prev => prev.map(n => n.id === data.note.id ? data.note : n));
    }).catch(() => { /* debounced save, fail silently */ });
  }, [selectedId, notes]);

  const handleTogglePin = async (note: Note) => {
    try {
      const data = await api.updateNote(note.id, { ...note, pinned: !note.pinned });
      setNotes(prev => prev.map(n => n.id === data.note.id ? data.note : n));
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await api.deleteNote(id);
      if (selectedId === id) setSelectedId(null);
      setNotes(prev => prev.filter(n => n.id !== id));
      setDeleteConfirmId(null);
    } catch (err: unknown) {
      if (err instanceof Error) setError(err.message);
    }
  };

  const handleTagAdd = async () => {
    if (!selectedNote) return;
    const tag = prompt('Enter tag name:');
    if (!tag || !tag.trim()) return;
    const newTags = [...(selectedNote.tags || []), tag.trim()];
    await handleUpdateField('tags', newTags);
  };

  const handleTagRemove = async (tag: string) => {
    if (!selectedNote) return;
    const newTags = (selectedNote.tags || []).filter(t => t !== tag);
    await handleUpdateField('tags', newTags);
  };

  const formatTime = (ts: string) => {
    const d = new Date(ts);
    const now = new Date();
    const diffMs = now.getTime() - d.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    const diffHrs = Math.floor(diffMins / 60);
    if (diffHrs < 24) return `${diffHrs}h ago`;
    const diffDays = Math.floor(diffHrs / 24);
    if (diffDays < 7) return `${diffDays}d ago`;
    return d.toLocaleDateString();
  };

  return (
    <div className="notes-page">
      <div className="notes-header">
        <h1 style={{ margin: 0 }}>Research Notes</h1>
        <button onClick={handleNewNote} className="notes-new-btn">
          + New Note
        </button>
      </div>

      {error && (
        <div className="notes-error">{error}</div>
      )}

      {allTags.length > 0 && (
        <div className="notes-tag-filter">
          <button
            className={`notes-tag-pill${filterTag === null ? ' active' : ''}`}
            onClick={() => setFilterTag(null)}
          >
            All
          </button>
          {allTags.map(tag => (
            <button
              key={tag}
              className={`notes-tag-pill${filterTag === tag ? ' active' : ''}`}
              onClick={() => setFilterTag(filterTag === tag ? null : tag)}
            >
              {tag}
            </button>
          ))}
        </div>
      )}

      <div className="notes-layout">
        <div className="notes-sidebar">
          {sortedGroups.length === 0 && (
            <p className="notes-empty">No notes yet. Create one to get started.</p>
          )}
          {sortedGroups.map(([group, groupNotes]) => (
            <div key={group} className="notes-group">
              <div className="notes-group-label">{group}</div>
              {groupNotes.map(note => (
                <div
                  key={note.id}
                  className={`notes-card${selectedId === note.id ? ' selected' : ''}`}
                  onClick={() => setSelectedId(note.id)}
                >
                  <div className="notes-card-header">
                    <span className="notes-card-title">{note.title || 'Untitled'}</span>
                    <button
                      className={`notes-pin-btn${note.pinned ? ' pinned' : ''}`}
                      onClick={(e) => { e.stopPropagation(); handleTogglePin(note); }}
                      title={note.pinned ? 'Unpin' : 'Pin'}
                    >
                      {note.pinned ? '\u2605' : '\u2606'}
                    </button>
                  </div>
                  <div className="notes-card-meta">
                    {note.symbol && <span className="notes-symbol-badge">{note.symbol}</span>}
                    {(note.tags || []).map(tag => (
                      <span key={tag} className="notes-tag-small">{tag}</span>
                    ))}
                    <span className="notes-card-time">{formatTime(note.updated_at)}</span>
                  </div>
                </div>
              ))}
            </div>
          ))}
        </div>

        <div className="notes-panel">
          {selectedNote ? (
            <>
              <div className="notes-panel-header">
                <input
                  className="notes-title-input"
                  value={selectedNote.title}
                  onChange={(e) => {
                    const val = e.target.value;
                    setNotes(prev => prev.map(n => n.id === selectedNote.id ? { ...n, title: val } : n));
                  }}
                  onBlur={() => handleUpdateField('title', selectedNote.title)}
                  placeholder="Note title"
                />
                <div className="notes-panel-actions">
                  <input
                    className="notes-symbol-input"
                    value={selectedNote.symbol}
                    onChange={(e) => {
                      const val = e.target.value.toUpperCase();
                      setNotes(prev => prev.map(n => n.id === selectedNote.id ? { ...n, symbol: val } : n));
                    }}
                    onBlur={() => handleUpdateField('symbol', selectedNote.symbol)}
                    placeholder="Symbol (e.g. AAPL)"
                  />
                  {deleteConfirmId === selectedNote.id ? (
                    <span className="notes-delete-confirm">
                      Delete?
                      <button onClick={() => handleDelete(selectedNote.id)} className="notes-delete-yes">Yes</button>
                      <button onClick={() => setDeleteConfirmId(null)} className="notes-delete-no">No</button>
                    </span>
                  ) : (
                    <button
                      onClick={() => setDeleteConfirmId(selectedNote.id)}
                      className="notes-delete-btn"
                      title="Delete note"
                    >
                      Delete
                    </button>
                  )}
                </div>
              </div>

              <div className="notes-tags-row">
                {(selectedNote.tags || []).map(tag => (
                  <span key={tag} className="notes-tag-pill removable">
                    {tag}
                    <button onClick={() => handleTagRemove(tag)} className="notes-tag-remove">&times;</button>
                  </span>
                ))}
                <button onClick={handleTagAdd} className="notes-tag-add">+ Tag</button>
              </div>

              <NoteEditor
                content={selectedNote.content}
                onUpdate={handleContentUpdate}
              />
            </>
          ) : (
            <div className="notes-empty-panel">
              <p>Select a note or create a new one to start writing.</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

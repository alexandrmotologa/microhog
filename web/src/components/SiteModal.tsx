import React, { useState, useEffect } from 'react';
import { SiteItem } from '../types/api';
import { X, Plus, Globe, Trash2, Check } from 'lucide-react';

interface SiteModalProps {
  currentSiteId: string;
  onSelectSite: (id: string) => void;
  onClose: () => void;
  onFetchSites: () => Promise<SiteItem[]>;
  onCreateSite: (id: string, name: string, domain: string) => Promise<void>;
  onDeleteSite: (id: string) => Promise<void>;
}

export const SiteModal: React.FC<SiteModalProps> = ({
  currentSiteId,
  onSelectSite,
  onClose,
  onFetchSites,
  onCreateSite,
  onDeleteSite,
}) => {
  const [sites, setSites] = useState<SiteItem[]>([]);
  const [id, setId] = useState<string>('');
  const [name, setName] = useState<string>('');
  const [domain, setDomain] = useState<string>('');
  const [error, setError] = useState<string | null>(null);

  const loadSites = async () => {
    try {
      const res = await onFetchSites();
      setSites(res || []);
    } catch (e: any) {
      setError(e.message);
    }
  };

  useEffect(() => {
    loadSites();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id.trim() || !name.trim()) return;
    try {
      setError(null);
      await onCreateSite(id.trim(), name.trim(), domain.trim());
      setId('');
      setName('');
      setDomain('');
      await loadSites();
    } catch (err: any) {
      setError(err.message || 'Failed to create site');
    }
  };

  const handleDelete = async (siteId: string) => {
    if (siteId === 'default') {
      alert('Cannot delete the default site.');
      return;
    }
    if (!confirm(`Delete site "${siteId}"?`)) return;
    try {
      await onDeleteSite(siteId);
      if (currentSiteId === siteId) {
        onSelectSite('default');
      }
      await loadSites();
    } catch (err: any) {
      alert(err.message || 'Failed to delete site');
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150">
        <div className="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-slate-950/60">
          <div className="flex items-center space-x-2">
            <Globe className="w-5 h-5 text-blue-400" />
            <h3 className="text-sm font-bold text-white">Manage Websites & Projects</h3>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-6 space-y-6">
          {/* Create new site form */}
          <form onSubmit={handleCreate} className="space-y-3 bg-slate-950 p-4 rounded-xl border border-slate-800">
            <div className="text-xs font-semibold text-slate-300">Add New Project</div>
            <div className="grid grid-cols-2 gap-2">
              <input
                type="text"
                placeholder="Site ID (e.g. blog)"
                value={id}
                onChange={(e) => setId(e.target.value)}
                className="bg-slate-900 border border-slate-700 text-slate-200 text-xs px-2.5 py-1.5 rounded outline-none focus:border-blue-500 font-mono"
                required
              />
              <input
                type="text"
                placeholder="Project Name (e.g. My Blog)"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="bg-slate-900 border border-slate-700 text-slate-200 text-xs px-2.5 py-1.5 rounded outline-none focus:border-blue-500"
                required
              />
            </div>
            <div className="flex gap-2">
              <input
                type="text"
                placeholder="Domain (optional, e.g. blog.mydomain.com)"
                value={domain}
                onChange={(e) => setDomain(e.target.value)}
                className="flex-1 bg-slate-900 border border-slate-700 text-slate-200 text-xs px-2.5 py-1.5 rounded outline-none focus:border-blue-500 font-mono"
              />
              <button
                type="submit"
                className="px-3 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded text-xs font-semibold flex items-center gap-1"
              >
                <Plus className="w-3.5 h-3.5" />
                Add
              </button>
            </div>
            {error && <div className="text-rose-400 text-xs">{error}</div>}
          </form>

          {/* Sites list */}
          <div className="space-y-2">
            <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
              Configured Sites ({sites.length})
            </div>
            <div className="space-y-1.5 max-h-56 overflow-y-auto pr-1">
              {sites.map((s) => (
                <div
                  key={s.id}
                  className={`flex items-center justify-between p-3 rounded-lg border text-xs cursor-pointer transition-colors ${
                    s.id === currentSiteId
                      ? 'bg-blue-600/10 border-blue-500/40 text-white'
                      : 'bg-slate-950/60 border-slate-800/80 text-slate-300 hover:border-slate-700'
                  }`}
                  onClick={() => {
                    onSelectSite(s.id);
                    onClose();
                  }}
                >
                  <div className="flex items-center space-x-2">
                    {s.id === currentSiteId && <Check className="w-4 h-4 text-blue-400" />}
                    <div>
                      <div className="font-semibold text-slate-200">{s.name}</div>
                      <div className="text-[11px] font-mono text-slate-500">
                        ID: {s.id} {s.domain && `&bull; ${s.domain}`}
                      </div>
                    </div>
                  </div>

                  {s.id !== 'default' && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDelete(s.id);
                      }}
                      className="p-1 text-slate-500 hover:text-red-400 transition-colors"
                      title="Delete site"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

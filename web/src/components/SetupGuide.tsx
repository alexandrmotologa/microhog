import React, { useState, useEffect } from 'react';
import { Copy, Check, Code, ShieldCheck, Terminal, Bell, Plus, Trash2, Send } from 'lucide-react';
import { WebhookItem } from '../types/api';

interface SetupGuideProps {
  currentSiteId: string;
  onFetchWebhooks: () => Promise<WebhookItem[]>;
  onSaveWebhook: (name: string, url: string, eventFilter: string) => Promise<WebhookItem>;
  onDeleteWebhook: (id: string) => Promise<void>;
  onTestWebhook: (url: string) => Promise<void>;
}

export const SetupGuide: React.FC<SetupGuideProps> = ({
  currentSiteId,
  onFetchWebhooks,
  onSaveWebhook,
  onDeleteWebhook,
  onTestWebhook,
}) => {
  const [siteId, setSiteId] = useState<string>(currentSiteId || 'default');
  const [recordReplays, setRecordReplays] = useState<boolean>(true);
  const [copiedSnippet, setCopiedSnippet] = useState<boolean>(false);
  const [copiedJs, setCopiedJs] = useState<boolean>(false);

  // Webhooks state
  const [webhooks, setWebhooks] = useState<WebhookItem[]>([]);
  const [whName, setWhName] = useState<string>('Discord Alerts');
  const [whUrl, setWhUrl] = useState<string>('');
  const [whFilter, setWhFilter] = useState<string>('all');
  const [testingUrl, setTestingUrl] = useState<string | null>(null);
  const [whMessage, setWhMessage] = useState<string | null>(null);

  const origin = typeof window !== 'undefined' ? window.location.origin : 'http://localhost:8080';

  const scriptTag = `<script 
  defer 
  src="${origin}/microhog.js" 
  data-site="${siteId}"${recordReplays ? '\n  data-record="true"' : ''}
  data-host="${origin}">
</script>`;

  const customEventCode = `// Dispatch a custom event anywhere in your app
window.microhog?.capture("plan_upgraded", {
  tier: "pro",
  annual: true,
  amount: 299
});`;

  const copyToClipboard = (text: string, isSnippet: boolean) => {
    navigator.clipboard.writeText(text);
    if (isSnippet) {
      setCopiedSnippet(true);
      setTimeout(() => setCopiedSnippet(false), 2000);
    } else {
      setCopiedJs(true);
      setTimeout(() => setCopiedJs(false), 2000);
    }
  };

  const loadWebhooks = async () => {
    try {
      const list = await onFetchWebhooks();
      setWebhooks(list || []);
    } catch {}
  };

  useEffect(() => {
    loadWebhooks();
  }, [currentSiteId]);

  const handleCreateWebhook = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!whUrl.trim()) return;
    try {
      setWhMessage(null);
      await onSaveWebhook(whName.trim(), whUrl.trim(), whFilter);
      setWhUrl('');
      await loadWebhooks();
      setWhMessage('Webhook registered successfully!');
      setTimeout(() => setWhMessage(null), 3000);
    } catch (err: any) {
      setWhMessage('Failed to save webhook: ' + err.message);
    }
  };

  const handleDeleteWebhook = async (id: string) => {
    if (!confirm('Remove this webhook?')) return;
    try {
      await onDeleteWebhook(id);
      setWebhooks((prev) => prev.filter((w) => w.id !== id));
    } catch {}
  };

  const handleTestWebhook = async (url: string) => {
    try {
      setTestingUrl(url);
      await onTestWebhook(url);
      alert('Test alert delivered successfully to your webhook!');
    } catch (err: any) {
      alert('Webhook verification failed: ' + err.message);
    } finally {
      setTestingUrl(null);
    }
  };

  return (
    <div className="space-y-6 max-w-4xl mx-auto">
      {/* Intro Card */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-6">
        <h2 className="text-base font-bold text-white flex items-center gap-2">
          <Code className="w-5 h-5 text-blue-400" />
          Integration & Tracking Snippet Generator
        </h2>
        <p className="text-xs text-slate-400 mt-1">
          Embed the lightweight 2KB tracker on your website. Operates without tracking cookies, requiring zero GDPR consent banners.
        </p>

        {/* Configuration options */}
        <div className="mt-5 grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label className="block text-xs font-semibold text-slate-300 mb-1.5">
              Site Identifier
            </label>
            <input
              type="text"
              value={siteId}
              onChange={(e) => setSiteId(e.target.value)}
              className="w-full bg-slate-950 border border-slate-700 text-slate-200 rounded-lg px-3 py-2 text-xs font-mono outline-none focus:border-blue-500"
              placeholder="default"
            />
          </div>

          <div className="flex items-center space-x-3 pt-6">
            <input
              type="checkbox"
              id="record-toggle"
              checked={recordReplays}
              onChange={(e) => setRecordReplays(e.target.checked)}
              className="w-4 h-4 rounded bg-slate-950 border-slate-700 text-blue-600 focus:ring-blue-500"
            />
            <label htmlFor="record-toggle" className="text-xs text-slate-300 cursor-pointer font-medium">
              Enable DOM Session Replays (<code className="text-blue-400 font-mono">data-record="true"</code>)
            </label>
          </div>
        </div>

        {/* Generated Snippet */}
        <div className="mt-5 relative">
          <div className="flex items-center justify-between bg-slate-950 px-4 py-2 border-t border-x border-slate-800 rounded-t-lg text-xs text-slate-400 font-mono">
            <span>HTML Snippet (place in &lt;head&gt;)</span>
            <button
              onClick={() => copyToClipboard(scriptTag, true)}
              className="flex items-center gap-1 text-slate-300 hover:text-white transition-colors"
            >
              {copiedSnippet ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
              <span>{copiedSnippet ? 'Copied' : 'Copy'}</span>
            </button>
          </div>
          <pre className="p-4 bg-slate-950 border border-slate-800 rounded-b-lg text-xs font-mono text-blue-300 overflow-x-auto leading-relaxed">
            {scriptTag}
          </pre>
        </div>
      </div>

      {/* Webhooks & Alerts Configuration */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-6">
        <h3 className="text-sm font-bold text-white flex items-center gap-2">
          <Bell className="w-4 h-4 text-purple-400" />
          Outbound Webhooks (Discord, Slack, or HTTP)
        </h3>
        <p className="text-xs text-slate-400 mt-1">
          Receive real-time notifications when critical events occur (e.g. conversions or rage clicks).
        </p>

        {/* Webhook Form */}
        <form onSubmit={handleCreateWebhook} className="mt-4 p-4 rounded-xl bg-slate-950 border border-slate-800 space-y-3">
          <div className="text-xs font-semibold text-slate-300">Add Notification Webhook</div>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
            <input
              type="text"
              placeholder="Name (e.g. Slack Team)"
              value={whName}
              onChange={(e) => setWhName(e.target.value)}
              className="bg-slate-900 border border-slate-700 text-slate-200 text-xs px-2.5 py-1.5 rounded outline-none focus:border-blue-500"
              required
            />
            <input
              type="url"
              placeholder="Webhook URL (Discord/Slack webhook URL)"
              value={whUrl}
              onChange={(e) => setWhUrl(e.target.value)}
              className="sm:col-span-2 bg-slate-900 border border-slate-700 text-slate-200 text-xs px-2.5 py-1.5 rounded outline-none focus:border-blue-500 font-mono"
              required
            />
          </div>

          <div className="flex items-center justify-between gap-3 pt-1">
            <div className="flex items-center space-x-2 text-xs">
              <span className="text-slate-400">Trigger on:</span>
              <select
                value={whFilter}
                onChange={(e) => setWhFilter(e.target.value)}
                className="bg-slate-900 border border-slate-700 text-slate-200 text-xs px-2 py-1 rounded outline-none"
              >
                <option value="all">All Events (except pageviews)</option>
                <option value="$rage_click">$rage_click (Frustration alerts)</option>
                <option value="signup_completed">signup_completed</option>
                <option value="click_signup">click_signup</option>
              </select>
            </div>

            <button
              type="submit"
              className="px-3 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded text-xs font-semibold flex items-center gap-1"
            >
              <Plus className="w-3.5 h-3.5" />
              Add Webhook
            </button>
          </div>

          {whMessage && <div className="text-blue-400 text-xs mt-1">{whMessage}</div>}
        </form>

        {/* Existing Webhooks List */}
        {webhooks.length > 0 && (
          <div className="mt-4 space-y-2">
            <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
              Active Webhooks ({webhooks.length})
            </div>
            {webhooks.map((w) => (
              <div
                key={w.id}
                className="flex items-center justify-between p-3 rounded-lg bg-slate-950/60 border border-slate-800 text-xs"
              >
                <div>
                  <div className="font-semibold text-slate-200 flex items-center gap-2">
                    {w.name}
                    <span className="text-[10px] font-mono px-1.5 py-0.2 rounded bg-slate-800 text-slate-400">
                      {w.event_filter}
                    </span>
                  </div>
                  <div className="text-[11px] font-mono text-slate-500 truncate max-w-md mt-0.5">
                    {w.url}
                  </div>
                </div>

                <div className="flex items-center space-x-2">
                  <button
                    onClick={() => handleTestWebhook(w.url)}
                    disabled={testingUrl === w.url}
                    className="flex items-center gap-1 px-2.5 py-1 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 text-[11px] transition-colors"
                  >
                    <Send className="w-3 h-3 text-blue-400" />
                    <span>{testingUrl === w.url ? 'Testing...' : 'Test'}</span>
                  </button>
                  <button
                    onClick={() => handleDeleteWebhook(w.id)}
                    className="p-1 text-slate-500 hover:text-red-400 transition-colors"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Custom Events Guide */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-6">
        <h3 className="text-sm font-bold text-white flex items-center gap-2">
          <Terminal className="w-4 h-4 text-emerald-400" />
          Custom Event & Goal Tracking
        </h3>
        <p className="text-xs text-slate-400 mt-1">
          Capture custom user interactions, conversions, or e-commerce purchases directly from JavaScript.
        </p>

        <div className="mt-4 relative">
          <div className="flex items-center justify-between bg-slate-950 px-4 py-2 border-t border-x border-slate-800 rounded-t-lg text-xs text-slate-400 font-mono">
            <span>JavaScript API</span>
            <button
              onClick={() => copyToClipboard(customEventCode, false)}
              className="flex items-center gap-1 text-slate-300 hover:text-white transition-colors"
            >
              {copiedJs ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
              <span>{copiedJs ? 'Copied' : 'Copy'}</span>
            </button>
          </div>
          <pre className="p-4 bg-slate-950 border border-slate-800 rounded-b-lg text-xs font-mono text-emerald-300 overflow-x-auto leading-relaxed">
            {customEventCode}
          </pre>
        </div>
      </div>

      {/* Privacy Masking Guide */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-6">
        <h3 className="text-sm font-bold text-white flex items-center gap-2">
          <ShieldCheck className="w-4 h-4 text-cyan-400" />
          Privacy Masking Rules
        </h3>
        <p className="text-xs text-slate-400 mt-1">
          Session replay automatically obscures sensitive fields before data leaves the visitor's browser:
        </p>
        <ul className="mt-3 space-y-2 text-xs text-slate-300 list-disc list-inside">
          <li>All <code className="text-blue-400 font-mono">&lt;input type="password"&gt;</code> fields are replaced with asterisks.</li>
          <li>Credit card inputs and inputs with <code className="text-blue-400 font-mono">autocomplete="cc-number"</code> are masked.</li>
          <li>Any element with the CSS class <code className="text-blue-400 font-mono">.microhog-mask</code> will have its text hidden.</li>
        </ul>
      </div>
    </div>
  );
};

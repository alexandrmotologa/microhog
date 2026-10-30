import React, { useState } from 'react';
import { Copy, Check, Code, ShieldCheck, Terminal } from 'lucide-react';

interface SetupGuideProps {
  currentSiteId: string;
}

export const SetupGuide: React.FC<SetupGuideProps> = ({ currentSiteId }) => {
  const [siteId, setSiteId] = useState<string>(currentSiteId || 'default');
  const [recordReplays, setRecordReplays] = useState<boolean>(true);
  const [copiedSnippet, setCopiedSnippet] = useState<boolean>(false);
  const [copiedJs, setCopiedJs] = useState<boolean>(false);

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

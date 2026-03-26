"use client";

const VERSION = "1.0.0";

export default function HelpPage() {
  return (
    <div className="flex-1 overflow-y-auto px-12 py-16">
      <div className="max-w-6xl mx-auto space-y-24">
        {/* Hero Section */}
        <section className="flex flex-col md:flex-row gap-16 items-start">
          <div className="flex-1">
            <h2 className="font-headline text-6xl font-bold tracking-tight leading-tight mb-6">
              Version <span className="text-[#ba9eff]">{VERSION}</span>
              <br />
              Local-First Intelligence.
            </h2>
            <p className="text-xl text-[#a3aac4] max-w-xl leading-relaxed font-light">
              Zoro is a privacy-centric AI companion built for high-performance knowledge
              synthesis. It lives on your machine, respects your data, and scales with your
              hardware.
            </p>
            <div className="mt-8 flex gap-4">
              <a
                href="https://github.com/urmzd/zoro"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 px-6 py-3 glass-panel border border-[#40485d]/20 rounded-xl hover:bg-[#1f2b49]/40 transition-all"
              >
                <span className="material-symbols-outlined">terminal</span>
                <span className="font-bold">v{VERSION}-stable</span>
              </a>
              <a
                href="https://github.com/urmzd/zoro"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 px-6 py-3 text-[#ba9eff] font-bold hover:translate-x-1 transition-transform"
              >
                <span>GitHub Repository</span>
                <span className="material-symbols-outlined">arrow_forward</span>
              </a>
            </div>
          </div>
          <div className="w-full md:w-1/3 aspect-square glass-panel rounded-3xl p-1 flex items-center justify-center relative overflow-hidden">
            <div className="absolute inset-0 bg-gradient-to-br from-[#ba9eff]/20 via-transparent to-[#699cff]/20 animate-pulse" />
            <span
              className="material-symbols-outlined text-8xl text-[#ba9eff] font-light"
              style={{ fontVariationSettings: "'FILL' 1" }}
            >
              cloud_off
            </span>
            <div className="absolute bottom-6 left-6 right-6 p-4 bg-[#192540]/80 backdrop-blur-md rounded-xl">
              <p className="text-xs font-bold text-[#ba9eff] uppercase tracking-widest mb-1">
                Status
              </p>
              <p className="text-sm">Fully Autonomous / Local Only</p>
            </div>
          </div>
        </section>

        {/* How it works */}
        <section className="space-y-8">
          <h3 className="font-headline text-3xl font-bold">How it works</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <FeatureCard
              icon="forum"
              iconColor="text-[#ba9eff]"
              title="Contextual Chat"
              description="Real-time conversations with access to your local files and previous conversation history without cloud latency."
            />
            <FeatureCard
              icon="travel_explore"
              iconColor="text-[#ff716a]"
              title="Deep Research"
              description="Automated web exploration using SearXNG to synthesize multi-source reports directly into your knowledge base."
            />
            <FeatureCard
              icon="account_tree"
              iconColor="text-[#699cff]"
              title="Knowledge Graph"
              description="Visualizing connections between concepts, files, and chat history powered by SurrealDB's relational logic."
            />
          </div>
        </section>

        {/* Services & Models */}
        <section className="grid grid-cols-1 md:grid-cols-2 gap-16">
          <div className="space-y-8">
            <h3 className="font-headline text-3xl font-bold">Integrated Services</h3>
            <div className="space-y-4">
              <ServiceCard
                icon="robot_2"
                iconColor="text-[#ba9eff]"
                name="Ollama"
                description="Inference Engine"
              />
              <ServiceCard
                icon="database"
                iconColor="text-[#ff716a]"
                name="SurrealDB"
                description="Multi-model Graph DB"
              />
              <ServiceCard
                icon="search"
                iconColor="text-[#699cff]"
                name="SearXNG"
                description="Privacy Meta-search"
              />
            </div>
          </div>
          <div className="space-y-8">
            <h3 className="font-headline text-3xl font-bold">Model Inventory</h3>
            <div className="bg-[#0f1930] rounded-3xl p-8 border border-[#40485d]/10">
              <ul className="space-y-6">
                <ModelBar name="qwen3.5:4b" role="Default Chat" color="bg-[#ba9eff]" width="w-full" />
                <ModelBar name="qwen3.5:0.8b" role="Fast Summary" color="bg-[#ff716a]" width="w-1/3" />
                <ModelBar name="nomic-embed-text" role="Embeddings" color="bg-[#699cff]" width="w-2/3" />
              </ul>
            </div>
          </div>
        </section>

        {/* Shortcuts */}
        <section className="space-y-8">
          <div className="flex items-end justify-between">
            <h3 className="font-headline text-3xl font-bold">Mastery Shortcuts</h3>
            <p className="text-[#a3aac4] text-sm">Efficiency is the only currency.</p>
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <ShortcutCard keys="Enter" label="Send Message" />
            <ShortcutCard keys="Shift + Enter" label="New Line" />
            <ShortcutCard keys="Ctrl + K" label="Global Search" />
            <ShortcutCard keys="Ctrl + N" label="New Session" />
          </div>
        </section>

        {/* Footer */}
        <footer className="pt-16 border-t border-[#40485d]/5 flex flex-col md:flex-row justify-between items-center gap-8 text-[#a3aac4] pb-12">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-[#ba9eff]/20 flex items-center justify-center">
              <span
                className="material-symbols-outlined text-[#ba9eff] text-sm"
                style={{ fontVariationSettings: "'FILL' 1" }}
              >
                star
              </span>
            </div>
            <span className="font-headline font-bold text-[#dee5ff]">Zoro OSS Project</span>
          </div>
          <div className="flex gap-8">
            <a
              href="https://github.com/urmzd/zoro"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-[#ba9eff] transition-colors flex items-center gap-2"
            >
              <span className="material-symbols-outlined text-lg">code</span>
              Source Code
            </a>
            <a
              href="https://github.com/urmzd/zoro/issues"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-[#ba9eff] transition-colors flex items-center gap-2"
            >
              <span className="material-symbols-outlined text-lg">bug_report</span>
              Report Issue
            </a>
          </div>
          <p className="text-xs">Apache License 2.0. Local first, always.</p>
        </footer>
      </div>
    </div>
  );
}

function FeatureCard({
  icon,
  iconColor,
  title,
  description,
}: {
  icon: string;
  iconColor: string;
  title: string;
  description: string;
}) {
  return (
    <div className="p-8 rounded-3xl bg-[#050a18] border border-[#40485d]/5 hover:border-[#ba9eff]/20 transition-all group">
      <span
        className={`material-symbols-outlined text-4xl ${iconColor} mb-6 group-hover:scale-110 transition-transform inline-block`}
      >
        {icon}
      </span>
      <h4 className="text-xl font-bold mb-3">{title}</h4>
      <p className="text-[#a3aac4] leading-relaxed text-sm">{description}</p>
    </div>
  );
}

function ServiceCard({
  icon,
  iconColor,
  name,
  description,
}: {
  icon: string;
  iconColor: string;
  name: string;
  description: string;
}) {
  return (
    <div className="flex items-center justify-between p-4 glass-panel rounded-2xl border border-[#40485d]/10">
      <div className="flex items-center gap-4">
        <div className={`w-10 h-10 rounded-full ${iconColor.replace("text-", "bg-")}/10 flex items-center justify-center`}>
          <span className={`material-symbols-outlined ${iconColor} text-xl`}>{icon}</span>
        </div>
        <div>
          <p className="font-bold">{name}</p>
          <p className="text-xs text-[#a3aac4]">{description}</p>
        </div>
      </div>
    </div>
  );
}

function ModelBar({
  name,
  role,
  color,
  width,
}: {
  name: string;
  role: string;
  color: string;
  width: string;
}) {
  return (
    <li className="flex flex-col gap-1">
      <div className="flex justify-between items-center">
        <span className={`font-bold ${color.replace("bg-", "text-")}`}>{name}</span>
        <span className="text-xs text-[#a3aac4]">{role}</span>
      </div>
      <div className="w-full h-1 bg-[#192540] rounded-full overflow-hidden">
        <div className={`${width} h-full ${color}`} />
      </div>
    </li>
  );
}

function ShortcutCard({ keys, label }: { keys: string; label: string }) {
  return (
    <div className="p-6 bg-black border border-[#40485d]/10 rounded-2xl flex flex-col items-center justify-center gap-4 hover:bg-[#050a18] transition-colors">
      <kbd className="px-3 py-1 bg-[#1f2b49] rounded-lg font-mono text-[#ba9eff] border-b-2 border-[#ba9eff]/40 text-sm">
        {keys}
      </kbd>
      <span className="text-xs text-[#a3aac4] uppercase tracking-widest">{label}</span>
    </div>
  );
}

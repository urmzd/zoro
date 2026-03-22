"use client";

import { IconBrandGithub, IconExternalLink } from "@tabler/icons-react";

const VERSION = "0.1.0";

export default function HelpPage() {
  return (
    <div className="flex-1 overflow-auto">
      <div className="max-w-2xl mx-auto px-6 py-8 space-y-8">
        {/* App identity */}
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">
            Zoro
          </h1>
          <p className="text-sm text-muted-foreground">
            Privacy-first research agent with a personal knowledge graph. v{VERSION}
          </p>
        </div>

        {/* What is Zoro */}
        <Section title="What is Zoro?">
          <p>
            Zoro is a local-first AI research assistant. It searches the web, extracts entities and
            relationships using local LLMs, and builds a persistent knowledge graph — all on your
            machine. Your data never leaves your device.
          </p>
        </Section>

        {/* How it works */}
        <Section title="How it works">
          <dl className="space-y-3">
            <DT title="Chat">
              Ask questions and get answers powered by local LLMs. The agent can search the web and
              your knowledge graph using tools.
            </DT>
            <DT title="Research">
              Submit a research query. Zoro searches the web, ingests results into the knowledge
              graph, discovers entities and relationships, then synthesizes a summary.
            </DT>
            <DT title="Knowledge Graph">
              Browse the accumulated knowledge as an interactive force-directed graph. Click nodes to
              see details and connections.
            </DT>
          </dl>
        </Section>

        {/* Services */}
        <Section title="Services">
          <p className="mb-2">
            Zoro manages these services automatically in desktop mode:
          </p>
          <ul className="space-y-1.5 text-sm">
            <li>
              <Strong>Ollama</Strong> — local LLM inference. Must be installed and running
              separately.
            </li>
            <li>
              <Strong>SurrealDB</Strong> — graph and session storage. Downloaded automatically on
              first launch.
            </li>
            <li>
              <Strong>SearXNG</Strong> — privacy-respecting web search. Installed into a local
              Python venv on first launch.
            </li>
          </ul>
          <p className="mt-2 text-xs text-muted-foreground">
            Check service status on the Logs page.
          </p>
        </Section>

        {/* Keyboard shortcuts */}
        <Section title="Shortcuts">
          <div className="grid grid-cols-2 gap-x-8 gap-y-1.5 text-sm">
            <KBD keys="Enter">Send message</KBD>
            <KBD keys="Shift + Enter">New line</KBD>
          </div>
        </Section>

        {/* Models */}
        <Section title="Models">
          <ul className="space-y-1.5 text-sm">
            <li>
              <Strong>qwen3.5:4b</Strong> — main reasoning model
            </li>
            <li>
              <Strong>qwen3.5:0.8b</Strong> — fast model for intent classification and autocomplete
            </li>
            <li>
              <Strong>nomic-embed-text</Strong> — text embeddings for the knowledge graph
            </li>
          </ul>
          <p className="mt-2 text-xs text-muted-foreground">
            Models can be changed via environment variables. See the README for details.
          </p>
        </Section>

        {/* Links */}
        <Section title="Links">
          <div className="flex flex-col gap-2">
            <ExtLink href="https://github.com/urmzd/zoro">
              <IconBrandGithub className="h-4 w-4" />
              Source code
            </ExtLink>
            <ExtLink href="https://github.com/urmzd/zoro/issues">
              <IconBrandGithub className="h-4 w-4" />
              Report an issue
            </ExtLink>
            <ExtLink href="https://ollama.ai">
              <IconExternalLink className="h-4 w-4" />
              Ollama
            </ExtLink>
          </div>
        </Section>

        {/* License */}
        <p className="text-xs text-muted-foreground pt-4 border-t border-border/50">
          Apache License 2.0. SearXNG is AGPL-3.0 — see THIRD-PARTY-LICENSES.md.
        </p>
      </div>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="space-y-2">
      <h2 className="text-sm font-semibold text-foreground">{title}</h2>
      <div className="text-sm text-muted-foreground leading-relaxed">{children}</div>
    </section>
  );
}

function DT({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <dt className="font-medium text-foreground">{title}</dt>
      <dd className="text-muted-foreground mt-0.5">{children}</dd>
    </div>
  );
}

function Strong({ children }: { children: React.ReactNode }) {
  return <span className="font-medium text-foreground">{children}</span>;
}

function KBD({ keys, children }: { keys: string; children: React.ReactNode }) {
  return (
    <>
      <span className="text-muted-foreground">{children}</span>
      <kbd className="text-xs font-mono px-1.5 py-0.5 rounded bg-muted border border-border text-foreground">
        {keys}
      </kbd>
    </>
  );
}

function ExtLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className="flex items-center gap-2 text-sm text-indigo-400 hover:text-indigo-300 transition-colors"
    >
      {children}
    </a>
  );
}

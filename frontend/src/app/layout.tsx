import type { Metadata } from "next";
import { Manrope, Space_Grotesk } from "next/font/google";
import { GlobalShortcuts } from "@/components/nav/global-shortcuts";
import { Sidebar } from "@/components/nav/sidebar";
import "./globals.css";

const spaceGrotesk = Space_Grotesk({
  variable: "--font-headline",
  subsets: ["latin"],
  display: "swap",
});

const manrope = Manrope({
  variable: "--font-body",
  subsets: ["latin"],
  display: "swap",
});

export const metadata: Metadata = {
  title: "Zoro — Luminous Intelligence",
  description: "Local-first AI research agent with persistent knowledge graph",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body
        className={`${spaceGrotesk.variable} ${manrope.variable} font-body antialiased selection:bg-primary/30`}
      >
        <div className="flex h-screen overflow-hidden">
          <Sidebar />
          <main className="flex-1 flex flex-col min-w-0 relative overflow-hidden">
            {/* Ambient background glows */}
            <div className="absolute -top-40 -right-40 w-[600px] h-[600px] bg-[#ba9eff]/3 rounded-full blur-[150px] pointer-events-none" />
            <div className="absolute -bottom-40 -left-40 w-[600px] h-[600px] bg-[#699cff]/3 rounded-full blur-[150px] pointer-events-none" />
            <GlobalShortcuts />
            <div className="flex-1 overflow-auto relative z-10">{children}</div>
          </main>
        </div>
      </body>
    </html>
  );
}

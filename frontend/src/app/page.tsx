import { SearchForm } from "@/components/research/search-form";
import { BackgroundBeams } from "@/components/ui/background-beams";

export default function Home() {
  return (
    <main className="relative flex min-h-screen flex-col items-center justify-center px-4">
      <BackgroundBeams />
      <div className="relative z-10 flex flex-col items-center gap-8 w-full max-w-3xl">
        <div className="text-center space-y-3">
          <h1 className="text-6xl font-bold tracking-tight bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">
            Zoro
          </h1>
          <p className="text-muted-foreground text-lg">
            AI-powered research with persistent knowledge
          </p>
        </div>
        <SearchForm />
      </div>
    </main>
  );
}

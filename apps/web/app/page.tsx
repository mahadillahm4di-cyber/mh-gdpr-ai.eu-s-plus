import Link from "next/link";

export default function LandingPage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center px-6">
      {/* Hero */}
      <div className="max-w-3xl text-center">
        <div className="mb-6 inline-block rounded-full border border-white/10 bg-white/5 px-4 py-1.5 text-sm text-white/60">
          Open Source &bull; GDPR-First &bull; Free
        </div>

        <h1 className="mb-6 text-5xl font-bold leading-tight tracking-tight md:text-7xl">
          Your AI Memory.
          <br />
          <span className="bg-gradient-to-r from-blue-400 via-orange-400 to-green-400 bg-clip-text text-transparent">
            Your Rules.
          </span>
        </h1>

        <p className="mx-auto mb-10 max-w-xl text-lg text-white/60">
          Switch between GPT, Claude, and Llama without losing context. Your
          memory stays on your machine. Visualize your AI brain in 3D.
        </p>

        {/* CTA Buttons */}
        <div className="flex flex-col items-center gap-4 sm:flex-row sm:justify-center">
          <Link
            href="/chat"
            className="rounded-lg bg-white px-8 py-3 text-lg font-semibold text-black transition hover:bg-white/90"
          >
            Try it free
          </Link>
          <Link
            href="/space"
            className="rounded-lg border border-white/20 px-8 py-3 text-lg font-semibold text-white transition hover:bg-white/10"
          >
            See your memory
          </Link>
        </div>
      </div>

      {/* Provider badges */}
      <div className="mt-20 flex items-center gap-8 text-sm text-white/40">
        <div className="flex items-center gap-2">
          <span
            className="h-2.5 w-2.5 rounded-full"
            style={{ backgroundColor: "#4A90D9" }}
          />
          OpenAI
        </div>
        <div className="flex items-center gap-2">
          <span
            className="h-2.5 w-2.5 rounded-full"
            style={{ backgroundColor: "#E87B35" }}
          />
          Anthropic
        </div>
        <div className="flex items-center gap-2">
          <span
            className="h-2.5 w-2.5 rounded-full"
            style={{ backgroundColor: "#4ADE80" }}
          />
          Ollama (Local)
        </div>
      </div>

      {/* Feature grid */}
      <div className="mt-24 grid max-w-4xl gap-6 md:grid-cols-3">
        <FeatureCard
          title="Universal Memory"
          description="Every conversation saved locally. Switch providers, keep everything."
        />
        <FeatureCard
          title="Zero Lock-in"
          description="GPT today, Claude tomorrow, Llama next week. Your data follows you."
        />
        <FeatureCard
          title="3D Brain View"
          description="See your memories as stars in space. Explore your AI brain visually."
        />
      </div>

      {/* Footer */}
      <footer className="mt-32 pb-8 text-sm text-white/30">
        mh-gdpr-ai.eu S+ &bull; Apache 2.0 &bull; Made by Mahadillah
      </footer>
    </main>
  );
}

function FeatureCard({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="rounded-xl border border-white/10 bg-white/5 p-6">
      <h3 className="mb-2 text-lg font-semibold">{title}</h3>
      <p className="text-sm text-white/50">{description}</p>
    </div>
  );
}

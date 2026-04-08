import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "mh-gdpr-ai.eu S+ — Your AI Memory, Your Rules",
  description:
    "Open-source protocol that unifies AI memory across all providers. Switch between GPT, Claude, and Llama without losing context. Visualize your AI brain in 3D.",
  openGraph: {
    title: "mh-gdpr-ai.eu S+",
    description: "Your AI Memory, Your Rules",
    type: "website",
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="dark">
      <body className="min-h-screen bg-black text-white antialiased">
        {children}
      </body>
    </html>
  );
}

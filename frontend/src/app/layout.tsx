import type { Metadata } from "next";
import "./globals.css";
import { AuthProvider } from "@/lib/auth";
import Header from "@/components/layout/Header";

export const metadata: Metadata = {
  title: "JobCrawler — AI-Powered Job Matching",
  description: "Upload your resume, get personalized job listings from LinkedIn, Naukri, Hirist, Instahyre, and more.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <AuthProvider>
          <Header />
          <main style={{ position: 'relative', zIndex: 1, minHeight: 'calc(100vh - 72px)' }}>
            {children}
          </main>
        </AuthProvider>
      </body>
    </html>
  );
}

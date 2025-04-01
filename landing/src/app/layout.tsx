import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Nixopus - Modern Infrastructure Management Platform",
  description: "Nixopus is a powerful infrastructure management platform that helps you deploy, manage, and scale your applications with ease. Built with modern technologies and designed for developers.",
  keywords: ["infrastructure", "devops", "cloud", "deployment", "management", "nix", "nixos", "kubernetes", "docker"],
  authors: [{ name: "Nixopus Team" }],
  creator: "Nixopus",
  publisher: "Nixopus",
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  metadataBase: new URL("https://nixopus.com"),
  alternates: {
    canonical: "/",
  },
  openGraph: {
    title: "Nixopus - Modern Infrastructure Management Platform",
    description: "Nixopus is a powerful infrastructure management platform that helps you deploy, manage, and scale your applications with ease.",
    url: "https://nixopus.com",
    siteName: "Nixopus",
    images: [
      {
        url: "/og-image.jpg",
        width: 1200,
        height: 630,
        alt: "Nixopus Platform Preview",
      },
    ],
    locale: "en_US",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Nixopus - Modern Infrastructure Management Platform",
    description: "Nixopus is a powerful infrastructure management platform that helps you deploy, manage, and scale your applications with ease.",
    images: ["/twitter-image.jpg"],
    creator: "@nixopus",
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
  verification: {
    google: "your-google-site-verification",
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" className="scroll-smooth">
      <head>
        <link rel="icon" href="/favicon.ico" />
        <link rel="apple-touch-icon" href="/apple-touch-icon.png" />
        <link rel="manifest" href="/site.webmanifest" />
        <meta name="theme-color" content="#000000" />
      </head>
      <body className={inter.className}>
        {children}
      </body>
    </html>
  );
}

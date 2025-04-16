import { NextResponse } from "next/server";

const DISCORD_WEBHOOK_URL = process.env.NEXT_PUBLIC_DISCORD_WEBHOOK_URL;

export async function POST(req: Request) {
  try {
    const { email } = await req.json();

    if (!email) {
      return NextResponse.json(
        { error: "Email is required" },
        { status: 400 }
      );
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      return NextResponse.json(
        { error: "Invalid email format" },
        { status: 400 }
      );
    }

    if (!DISCORD_WEBHOOK_URL) {
      console.error("Discord webhook URL not configured");
      return NextResponse.json(
        { error: "Server configuration error" },
        { status: 500 }
      );
    }

    const response = await fetch(DISCORD_WEBHOOK_URL, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        embeds: [{
          title: "🎉 New Waitlist Signup!",
          description: `Someone just joined the waitlist:\n**Email:** ${email}`,
          color: 0x6366f1, 
          timestamp: new Date().toISOString(),
          footer: {
            text: "Nixopus Waitlist"
          }
        }],
      }),
    });

    if (!response.ok) {
      throw new Error("Failed to send Discord notification");
    }

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error("Error processing waitlist signup:", error);
    return NextResponse.json(
      { error: "Error processing waitlist signup" },
      { status: 500 }
    );
  }
} 
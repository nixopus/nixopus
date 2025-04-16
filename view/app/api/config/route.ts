import { NextResponse } from 'next/server';

export async function GET() {
  return NextResponse.json({
    baseUrl: process.env.API_URL || 'http://localhost:8080/api',
    websocketUrl: process.env.WEBSOCKET_URL || 'ws://localhost:8080/ws',
    webhookUrl: process.env.WEBHOOK_URL || 'http://localhost:8080/webhook',
    port: process.env.NEXT_PUBLIC_PORT || '7443'
  });
}

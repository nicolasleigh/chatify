import { auth } from "@clerk/nextjs/server";
import { NextRequest, NextResponse } from "next/server";
import { AccessToken } from "livekit-server-sdk";
import { baseUrl } from "@/api/utils";

// Do not cache endpoint result
export const revalidate = 0;

export async function GET(req: NextRequest) {
  const { userId, getToken } = await auth();
  if (!userId) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const room = req.nextUrl.searchParams.get("room");
  if (!room) {
    return NextResponse.json({ error: 'Missing "room" query parameter' }, { status: 400 });
  } else if (!/^\d+$/.test(room)) {
    return NextResponse.json({ error: "Invalid room" }, { status: 400 });
  }

  const clerkToken = await getToken();
  if (!clerkToken) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const accessResponse = await fetch(`${baseUrl}/messages/${room}`, {
    headers: {
      Authorization: `Bearer ${clerkToken}`,
    },
    cache: "no-store",
  });
  if (!accessResponse.ok) {
    return NextResponse.json({ error: "Conversation access denied" }, { status: 403 });
  }

  const apiKey = process.env.LIVEKIT_API_KEY;
  const apiSecret = process.env.LIVEKIT_API_SECRET;
  const wsUrl = process.env.NEXT_PUBLIC_LIVEKIT_URL;

  if (!apiKey || !apiSecret || !wsUrl) {
    return NextResponse.json({ error: "Server misconfigured" }, { status: 500 });
  }

  const at = new AccessToken(apiKey, apiSecret, { identity: userId });
  at.addGrant({ room, roomJoin: true, canPublish: true, canSubscribe: true });

  return NextResponse.json({ token: await at.toJwt() }, { headers: { "Cache-Control": "no-store" } });
}

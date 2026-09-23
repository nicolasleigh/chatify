import { Webhook } from "svix";
import { headers } from "next/headers";
import { WebhookEvent } from "@clerk/nextjs/server";

// Clerk Webhook: https://clerk.com/docs/webhooks/sync-data
export async function POST(req: Request) {
  const CLERK_WEBHOOK_SECRET = process.env.CLERK_WEBHOOK_SECRET;

  if (!CLERK_WEBHOOK_SECRET) {
    throw new Error("Error: Please add CLERK_WEBHOOK_SECRET from Clerk Dashboard to .env or .env.local");
  }

  // Create new Svix instance with secret
  const wh = new Webhook(CLERK_WEBHOOK_SECRET);

  // Get headers
  const headerPayload = await headers();
  const svix_id = headerPayload.get("svix-id");
  const svix_timestamp = headerPayload.get("svix-timestamp");
  const svix_signature = headerPayload.get("svix-signature");

  // If there are no headers, error out
  if (!svix_id || !svix_timestamp || !svix_signature) {
    return new Response("Error: Missing Svix headers", {
      status: 400,
    });
  }

  // Get body
  const payload = await req.json();
  const body = JSON.stringify(payload);

  let evt: WebhookEvent;

  // Verify payload with headers
  try {
    evt = wh.verify(body, {
      "svix-id": svix_id,
      "svix-timestamp": svix_timestamp,
      "svix-signature": svix_signature,
    }) as WebhookEvent;
  } catch (err) {
    console.error("Error: Could not verify webhook:", err);
    return new Response("Error: Verification error", {
      status: 400,
    });
  }

  if (!evt) {
    return new Response("Could not validate Clerk payload", { status: 400 });
  }
  switch (evt.type) {
    case "user.created": // intentional fallthrough
    case "user.updated":
      console.log("Creating/Updating User:", evt.data.id);

      await fetch("https://back.chat.linze.pro/user", {
        method: "POST",
        body: JSON.stringify({
          username: `${evt.data.first_name} ${evt.data.last_name}`,
          image_url: evt.data.image_url,
          clerk_id: evt.data.id,
          email: evt.data.email_addresses[0].email_address,
        }),
        headers: {
          "Content-Type": "application/json",
          "X-Internal-Webhook-Secret": process.env.INTERNAL_WEBHOOK_SECRET ?? "",
        },
      });

      break;

    default:
      console.log("Ignored Clerk webhook event", evt.type);
  }

  return new Response("Webhook received", { status: 200 });
}

import type { PushSubscriptionPayload } from "@/api/notifications";

const pushServiceWorkerPath = "/push-sw.js";
const pushServiceWorkerScope = "/push/";

function decodeBase64Url(value: string): Uint8Array {
  const padding = "=".repeat((4 - (value.length % 4)) % 4);
  const normalized = `${value}${padding}`.replace(/-/g, "+").replace(/_/g, "/");
  const decoded = window.atob(normalized);
  return Uint8Array.from(decoded, (character) => character.charCodeAt(0));
}

export function isWebPushSupported(): boolean {
  return (
    typeof window !== "undefined" &&
    "serviceWorker" in navigator &&
    "PushManager" in window &&
    "Notification" in window
  );
}

export async function getOrCreateWebPushSubscription(vapidPublicKey: string): Promise<PushSubscription> {
  if (!isWebPushSupported()) {
    throw new Error("This browser does not support Web Push notifications");
  }
  if (!vapidPublicKey.trim()) {
    throw new Error("Web Push is not configured");
  }

  const registration = await navigator.serviceWorker.register(pushServiceWorkerPath, {
    scope: pushServiceWorkerScope,
  });
  const existingSubscription = await registration.pushManager.getSubscription();
  if (existingSubscription) {
    return existingSubscription;
  }

  return registration.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: decodeBase64Url(vapidPublicKey),
  });
}

export function serializeWebPushSubscription(subscription: PushSubscription): PushSubscriptionPayload {
  const serialized = subscription.toJSON();
  const p256dh = serialized.keys?.p256dh;
  const auth = serialized.keys?.auth;
  if (!serialized.endpoint || !p256dh || !auth) {
    throw new Error("Browser returned an incomplete Web Push subscription");
  }

  return {
    endpoint: serialized.endpoint,
    keys: { p256dh, auth },
  };
}

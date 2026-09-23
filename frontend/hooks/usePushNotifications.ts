"use client";

import { createPushSubscription } from "@/api/notifications";
import { getOrCreateWebPushSubscription, isWebPushSupported, serializeWebPushSubscription } from "@/lib/webPush";
import { useAuth } from "@clerk/nextjs";
import { useCallback, useEffect, useState } from "react";

export type PushNotificationStatus = "unavailable" | "idle" | "enabling" | "enabled" | "denied" | "error";

const vapidPublicKey = process.env.NEXT_PUBLIC_WEB_PUSH_VAPID_PUBLIC_KEY || "";

export default function usePushNotifications() {
  const { getToken } = useAuth();
  const [status, setStatus] = useState<PushNotificationStatus>("idle");

  useEffect(() => {
    if (!vapidPublicKey || !isWebPushSupported()) {
      setStatus("unavailable");
      return;
    }

    if (Notification.permission === "denied") {
      setStatus("denied");
      return;
    }

    let cancelled = false;
    navigator.serviceWorker
      .getRegistration("/push/")
      .then((registration) => registration?.pushManager.getSubscription())
      .then((subscription) => {
        if (!cancelled && subscription) {
          setStatus("enabled");
        }
      })
      .catch(() => {
        if (!cancelled) {
          setStatus("idle");
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const enable = useCallback(async () => {
    if (!vapidPublicKey || !isWebPushSupported()) {
      setStatus("unavailable");
      return;
    }

    setStatus("enabling");
    try {
      const permission = await Notification.requestPermission();
      if (permission !== "granted") {
        setStatus(permission === "denied" ? "denied" : "idle");
        return;
      }

      const token = await getToken();
      if (!token) {
        throw new Error("Authentication token is unavailable");
      }

      const subscription = await getOrCreateWebPushSubscription(vapidPublicKey);
      await createPushSubscription({
        subscription: serializeWebPushSubscription(subscription),
        token,
        deviceLabel: "Web browser",
      });
      setStatus("enabled");
    } catch {
      setStatus("error");
    }
  }, [getToken]);

  return { status, enable };
}

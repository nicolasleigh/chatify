"use client";

import usePushNotifications from "@/hooks/usePushNotifications";
import { Bell, BellOff, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function PushNotificationButton() {
  const { status, enable } = usePushNotifications();

  if (status === "unavailable") {
    return null;
  }

  const isBusy = status === "enabling";
  const isEnabled = status === "enabled";
  const isDenied = status === "denied";
  const label = isEnabled
    ? "Notifications enabled"
    : isDenied
      ? "Notifications blocked in browser settings"
      : status === "error"
        ? "Retry notifications"
        : "Enable offline notifications";

  return (
    <Button
      aria-label={label}
      disabled={isBusy || isDenied}
      onClick={enable}
      size='icon'
      title={label}
      variant={isEnabled ? "default" : "outline"}
    >
      {isBusy ? <Loader2 className='animate-spin' /> : isDenied ? <BellOff /> : <Bell />}
    </Button>
  );
}

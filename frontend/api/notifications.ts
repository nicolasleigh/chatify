import { baseUrl } from "./utils";

export type PushSubscriptionPayload = {
  endpoint: string;
  keys: {
    p256dh: string;
    auth: string;
  };
};

type PushSubscriptionResponse = {
  id: number;
  endpoint: string;
  enabled: boolean;
};

export async function createPushSubscription({
  subscription,
  token,
  deviceLabel,
}: {
  subscription: PushSubscriptionPayload;
  token: string;
  deviceLabel?: string;
}): Promise<PushSubscriptionResponse> {
  const response = await fetch(`${baseUrl}/notifications/subscriptions`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({
      endpoint: subscription.endpoint,
      keys: subscription.keys,
      device_label: deviceLabel,
    }),
  });

  if (!response.ok) {
    throw new Error(`Unable to save notification subscription (${response.status})`);
  }

  return (await response.json()) as PushSubscriptionResponse;
}

export async function disablePushSubscription({ subscriptionId, token }: { subscriptionId: number; token: string }) {
  const response = await fetch(`${baseUrl}/notifications/subscriptions/${subscriptionId}`, {
    method: "DELETE",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    throw new Error(`Unable to disable notification subscription (${response.status})`);
  }
}

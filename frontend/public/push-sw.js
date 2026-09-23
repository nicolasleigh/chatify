/*
 * This worker is intentionally separate from the existing root Workbox
 * service worker. The root worker owns application caching, while this worker
 * owns Push events and notification clicks under the /push/ scope.
 */
self.addEventListener("push", (event) => {
  let payload = {};

  try {
    payload = event.data ? event.data.json() : {};
  } catch {
    payload = { body: "You have a new message" };
  }

  const title = typeof payload.title === "string" && payload.title.trim()
    ? payload.title
    : "New message";
  const body = typeof payload.body === "string" && payload.body.trim()
    ? payload.body
    : "You have a new message";
  const conversationId = Number(payload.conversation_id);

  event.waitUntil(
    self.registration.showNotification(title, {
      body,
      icon: "/icon512_rounded.png",
      badge: "/icon512_maskable.png",
      tag: Number.isInteger(conversationId) && conversationId > 0
        ? `chatify-conversation-${conversationId}`
        : "chatify-message",
      data: {
        conversationId: Number.isInteger(conversationId) && conversationId > 0
          ? conversationId
          : null,
      },
    }),
  );
});

self.addEventListener("notificationclick", (event) => {
  event.notification.close();

  const conversationId = event.notification.data?.conversationId;
  const targetPath = Number.isInteger(conversationId) && conversationId > 0
    ? `/conversations/${conversationId}`
    : "/conversations";

  event.waitUntil(
    self.clients.matchAll({ type: "window", includeUncontrolled: true }).then((clients) => {
      for (const client of clients) {
        if ("focus" in client) {
          client.navigate(targetPath);
          return client.focus();
        }
      }

      if (self.clients.openWindow) {
        return self.clients.openWindow(targetPath);
      }

      return undefined;
    }),
  );
});

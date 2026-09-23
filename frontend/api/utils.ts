let baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8084";
let wsUrl = process.env.NEXT_PUBLIC_WS_BASE_URL || "ws://localhost:8084";

if (process.env.NODE_ENV === "production") {
  baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || "https://back.chat.linze.pro";
  wsUrl = process.env.NEXT_PUBLIC_WS_BASE_URL || "wss://back.chat.linze.pro";
}

export { baseUrl, wsUrl };

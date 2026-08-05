import { z } from "zod";
import { authHeaders, baseUrl } from "./utils";

type getConversationParams = {
  conversation_id: number;
  token: string;
};
const conversationSchema = z.array(
  z.object({
    current_user_id: z.number(),
    other_member_id: z.number(),
    other_member_username: z.string(),
    other_member_email: z.string(),
    other_member_image_url: z.string(),
    other_member_last_seen_message_id: z.nullable(z.number()),
    conversation_id: z.number(),
    conversation_name: z.nullable(z.string()),
    is_group: z.boolean(),
    last_message_id: z.nullable(z.number()),
    unseen_message_count: z.nullable(z.number()),
  })
);
export async function getConversation({
  conversation_id,
  token,
}: getConversationParams): Promise<z.infer<typeof conversationSchema>> {
  const response = await fetch(`${baseUrl}/conversation/${conversation_id}`, {
    method: "GET",
    headers: authHeaders(token),
  });
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
  const data = await response.json();
  if (!data || data.length === 0) {
    return [];
  }
  const conv = conversationSchema.parse(data);
  return conv;
}

type getAllConversationsParams = {
  token: string;
};
const allConversationsSchema = z.array(conversationSchema);
export async function getAllConversations({
  token,
}: getAllConversationsParams): Promise<z.infer<typeof allConversationsSchema>> {
  const response = await fetch(`${baseUrl}/conversations`, {
    method: "GET",
    headers: authHeaders(token),
  });
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
  let data = await response.json();
  if (!data || data.length === 0) {
    return [];
  }
  try {
    data = allConversationsSchema.parse(data);
  } catch (error) {
    console.error("Schema validation error:", error);
  }

  return data;
}

type createGroupParams = {
  member_id_arr: number[];
  name: string;
  token: string;
};
export async function createGroup({ member_id_arr, name, token }: createGroupParams) {
  const response = await fetch(`${baseUrl}/group/create`, {
    method: "POST",
    headers: {
      ...authHeaders(token),
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      member_id_arr,
      name,
    }),
  });
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
}

type leaveGroupParams = {
  conversation_id: number;
  token: string;
};
export async function leaveGroup({ conversation_id, token }: leaveGroupParams) {
  const response = await fetch(`${baseUrl}/group/leave/${conversation_id}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
}

type deleteGroupParams = {
  conversation_id: number;
  token: string;
};
export async function deleteGroup({ conversation_id, token }: deleteGroupParams) {
  const response = await fetch(`${baseUrl}/group/delete/${conversation_id}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
}

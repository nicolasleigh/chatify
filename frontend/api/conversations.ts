import { z } from "zod";
import { baseUrl } from "./utils";

type getConversationParams = {
  clerk_id: string;
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
  clerk_id,
  conversation_id,
  token,
}: getConversationParams): Promise<z.infer<typeof conversationSchema>> {
  const response = await fetch(`${baseUrl}/conversation/${clerk_id}/${conversation_id}`, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
    },
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
  clerk_id: string;
  token: string;
};
const allConversationsSchema = z.array(conversationSchema);
export async function getAllConversations({
  clerk_id,
  token,
}: getAllConversationsParams): Promise<z.infer<typeof allConversationsSchema>> {
  const response = await fetch(`${baseUrl}/conversations/${clerk_id}`, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
    },
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
  clerk_id: string;
  token: string;
};
export async function createGroup({ member_id_arr, name, clerk_id, token }: createGroupParams) {
  const response = await fetch(`${baseUrl}/group/create/${clerk_id}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
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
  clerk_id: string;
  conversation_id: number;
  token: string;
};
export async function leaveGroup({ clerk_id, conversation_id, token }: leaveGroupParams) {
  const response = await fetch(`${baseUrl}/group/leave/${clerk_id}/${conversation_id}`, {
    method: "DELETE",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
}

type deleteGroupParams = {
  clerk_id: string;
  conversation_id: number;
  token: string;
};
export async function deleteGroup({ clerk_id, conversation_id, token }: deleteGroupParams) {
  const response = await fetch(`${baseUrl}/group/delete/${clerk_id}/${conversation_id}`, {
    method: "DELETE",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
}

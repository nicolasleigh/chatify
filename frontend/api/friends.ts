import { z } from "zod";
import { baseUrl } from "./utils";

type createRequestParams = {
  email: string;
  token: string;
};
export async function createRequest({ email, token }: createRequestParams) {
  const response = await fetch(`${baseUrl}/request`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({ email }),
  });
  const data = await response.json();
  if (!response.ok) {
    if (response.status === 404 || 403) {
      throw new Error(data);
    }
    throw new Error("Unexpected error occurs.");
  }
}

type denyRequestParams = {
  request_id: number;
  token: string;
};
export async function denyRequest({ request_id, token }: denyRequestParams) {
  // When using react-query, be careful in wrapping in try-catch block, if you doing so, onError may not be triggered
  const response = await fetch(`${baseUrl}/deny/${request_id}`, {
    method: "DELETE",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  const data = await response.json();
  console.log(data);
}

type acceptRequestParams = {
  request_id: number;
  token: string;
};
export async function acceptRequest({ request_id, token }: acceptRequestParams) {
  const response = await fetch(`${baseUrl}/request/accept/${request_id}`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
}

const getFriendsSchema = z.array(
  z.object({
    id: z.number(),
    username: z.string(),
    email: z.string(),
    clerk_id: z.string(),
    image_url: z.string(),
  })
);
type getFriendsParams = {
  clerk_id: string;
  token: string;
};
export async function getFriends({ clerk_id, token }: getFriendsParams): Promise<z.infer<typeof getFriendsSchema>> {
  const response = await fetch(`${baseUrl}/friends/${clerk_id}`, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
  let data = await response.json();
  data = getFriendsSchema.parse(data);
  return data;
}

type deleteFriendParams = {
  conversation_id: number;
  token: string;
};
export async function deleteFriend({ conversation_id, token }: deleteFriendParams) {
  const response = await fetch(`${baseUrl}/friend/${conversation_id}`, {
    method: "DELETE",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }
}

type getRequestsParams = {
  clerk_id: string;
  token: string;
};
const requestsSchema = z.array(
  z.object({
    id: z.number(),
    username: z.string(),
    image_url: z.string(),
    email: z.string(),
    request_count: z.number(),
    sender_id: z.number(),
    receiver_id: z.number(),
  })
);
export async function getRequests({ clerk_id, token }: getRequestsParams): Promise<z.infer<typeof requestsSchema>> {
  const response = await fetch(`${baseUrl}/requests/${clerk_id}`, {
    method: "GET",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  const data = await response.json();
  if (!data) {
    return [];
  }
  const friendRequests = requestsSchema.parse(data);
  return friendRequests;
}

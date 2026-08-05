"use client";

import ConversationFallback from "@/components/shared/conversation/ConversationFallback";
import ItemList from "@/components/shared/item-list/ItemList";
import AddFriendDialog from "./_components/AddFriendDialog";
import { Loader2 } from "lucide-react";
import Request from "./_components/Request";
import { useQuery } from "@tanstack/react-query";
import { getRequests } from "@/api/friends";
import { useAuth } from "@clerk/nextjs";

export default function FriendsPage() {
  const { getToken } = useAuth();
  const { data: requests } = useQuery({
    queryKey: ["friend_requests"],
    queryFn: async () => {
      const token = await getToken();
      if (!token) {
        throw new Error("User token is not available");
      }
      return getRequests({ token });
    },
  });
  return (
    <>
      <ItemList title='Friends' action={<AddFriendDialog />}>
        {requests ? (
          requests.length === 0 ? (
            <p className='w-full h-full flex items-center justify-center'>No friend requests found</p>
          ) : (
            requests.map((req) => {
              return (
                <Request
                  key={req.id}
                  id={req.id}
                  imageUrl={req.image_url}
                  username={req.username}
                  email={req.email}
                />
              );
            })
          )
        ) : (
          <Loader2 className='h-8 w-8 animate-spin' />
        )}
      </ItemList>
      <ConversationFallback />
    </>
  );
}

"use client";

import { getConversation } from "@/api/conversations";
import ConversationContainer from "@/components/shared/conversation/ConversationContainer";
import { useAuthInfo } from "@/hooks/useAuthInfo";
import { useMessagesQuery } from "@/hooks/useMessagesQuery";
import { useQuery } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import React, { useEffect, useState } from "react";
import Header from "./_components/Header";
import Body from "./_components/body/Body";
import DeleteGroupDialog from "./_components/dialogs/DeleteGroupDialog";
import LeaveGroupDialog from "./_components/dialogs/LeaveGroupDialog";
import RemoveFriendDialog from "./_components/dialogs/RemoveFriendDialog";
import ChatInput from "./_components/input/ChatInput";

type Props = {
  params: Promise<{ conversationId: string }>;
};

export default function ConversationPage({ params }: Props) {
  const { conversationId } = React.use(params);
  const conversationNumber = Number(conversationId);
  const { userId: clerk_id, token } = useAuthInfo();
  const { data: conversation } = useQuery({
    queryKey: ["conversation", conversationId],
    queryFn: () => {
      if (!clerk_id) {
        throw new Error("Clerk user not found");
      }
      if (!token) {
        throw new Error("Authentication is not ready");
      }
      return getConversation({ conversation_id: conversationNumber, clerk_id, token });
    },
    enabled: Boolean(clerk_id && token),
  });

  // const { addMessage, isAddMessagePending, isLoading, isReady, messages: storeMessages } = useMessagesStore({ userId });

  // const { data: messages, isSuccess } = useQuery({
  //   queryKey: ["messages", conversationId],
  //   queryFn: () => {
  //     if (!token) {
  //       throw new Error("User token not found");
  //     }
  //     return getMessages(conversationId, token);
  //   },
  //   // gcTime: 24 * 60 * 60 * 1000, // 24H
  //   // retry: false,
  // });

  // if (isSuccess) {
  //   // messages.map((message) => {});
  //   // console.log(messages);
  //   storeDataInIndexedDB(clerk_id, conversationId, messages);
  // }

  // useEffect(() => {
  //   const fetchFromIndexedDB = async () => {
  //     const data = await getDataFromIndexedDB(userId, conversationId);
  //     console.log("getDataFromIndexedDB", data);
  //     setMessages(data);
  //   };
  //   fetchFromIndexedDB();
  // }, [userId, conversationId, msg]);

  // console.log("messages", msg);

  const { messages, websocket } = useMessagesQuery(conversationId, token || "");

  const [removeFriendDialogOpen, setRemoveFriendDialogOpen] = useState(false);
  const [deleteGroupDialogOpen, setDeleteGroupDialogOpen] = useState(false);
  const [leaveGroupDialogOpen, setLeaveGroupDialogOpen] = useState(false);
  const [callType, setCallType] = useState<"audio" | "video" | null>(null);

  // const websocket = useWebsocket({
  //   url: `${wsUrl}/ws/${conversationId}`,
  //   token: token || "", // Pass an empty string if token is not available
  // });

  useEffect(() => {
    if (!token) {
      console.log("No token available, WebSocket not connected");
    }
  }, [token]);

  return conversation === undefined ? (
    <div className='w-full h-full flex items-center justify-center'>
      <Loader2 className='h-8 w-8 animate-spin' />
    </div>
  ) : conversation === null ? (
    <p className='w-full h-full flex items-center justify-center'>Conversation not found</p>
  ) : (
    <ConversationContainer>
      <RemoveFriendDialog
        conversationId={conversationNumber}
        open={removeFriendDialogOpen}
        setOpen={setRemoveFriendDialogOpen}
      />
      <DeleteGroupDialog
        conversationId={conversationNumber}
        open={deleteGroupDialogOpen}
        setOpen={setDeleteGroupDialogOpen}
        clerkId={clerk_id || ""}
      />
      <LeaveGroupDialog
        conversationId={conversationNumber}
        open={leaveGroupDialogOpen}
        setOpen={setLeaveGroupDialogOpen}
      />
      <Header
        name={
          (conversation[0]?.is_group ? conversation[0].conversation_name : conversation[0].other_member_username) || ""
        }
        imageUrl={conversation[0].is_group ? undefined : conversation[0].other_member_image_url}
        options={
          conversation[0].is_group
            ? [
                { label: "Leave group", destructive: false, onClick: () => setLeaveGroupDialogOpen(true) },
                { label: "Delete group", destructive: true, onClick: () => setDeleteGroupDialogOpen(true) },
              ]
            : [{ label: "Remove friend", destructive: true, onClick: () => setRemoveFriendDialogOpen(true) }]
        }
        setCallType={setCallType}
      />
      <Body
        members={conversation}
        callType={callType}
        setCallType={setCallType}
        currentUserId={conversation[0].current_user_id}
        msg={messages}
      />
      <ChatInput sender_id={conversation[0].current_user_id} websocket={websocket} />
    </ConversationContainer>
  );
}

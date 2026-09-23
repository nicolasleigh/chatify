"use client";

import { Button } from "@/components/ui/button";
import { api } from "@/convex/_generated/api";
import useConversation from "@/hooks/useConversation";
import useMutationState from "@/hooks/useMutationState";
import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { ControlBar, LiveKitRoom, VideoConference } from "@livekit/components-react";
import "@livekit/components-styles";

type Props = {
  video: boolean;
  audio: boolean;
  handleDisconnect: () => void;
};

export default function CallRoom({ video, audio, handleDisconnect }: Props) {
  const [token, setToken] = useState("");
  const { conversationId } = useConversation();
  const { mutate: createMessage } = useMutationState(api.message.create);

  useEffect(() => {
    (async () => {
      try {
        const res = await fetch(`/api/livekit?room=${conversationId}`);
        const data = await res.json();

        if (!res.ok) {
          throw new Error(data.error || "Could not join the call");
        }

        setToken(data.token);
      } catch (error) {
        toast.error("Could not join the call");
        console.log(error);
      }
    })();
  }, [conversationId]);

  if (token === "") {
    return (
      <div className='w-full h-full flex flex-col justify-center items-center'>
        <Loader2 className='animate-spin h-16 w-16 text-foreground' />
        <p className='text-sm text-foreground'>Joining call...</p>
        <Button className='mt-4' variant='destructive' onClick={handleDisconnect}>
          Cancel
        </Button>
      </div>
    );
  }
  return (
    <div className='w-full h-full'>
      <LiveKitRoom
        data-lk-theme='default'
        serverUrl={process.env.NEXT_PUBLIC_LIVEKIT_URL}
        token={token}
        connect={true}
        video={video}
        audio={audio}
        onDisconnected={() => handleDisconnect()}
        onConnected={() => {
          createMessage({ conversationId, type: "call", content: [] });
        }}
      >
        <VideoConference />
        <ControlBar />
      </LiveKitRoom>
    </div>
  );
}

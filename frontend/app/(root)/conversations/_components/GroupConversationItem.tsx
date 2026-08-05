import { getConversationLastMessage } from "@/api/messages";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import useGetUnseenCount from "@/hooks/useGetUnseenCount";
import { useAuth } from "@clerk/nextjs";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";

type Props = {
  id: number;
  name: string;
  // unseenCount: number;
  lastMessageId: number | null;
};

export default function GroupConversationItem({ id, name, lastMessageId: message_id }: Props) {
  const { getToken } = useAuth();
  const unseenCnt = useGetUnseenCount();
  const cnt = unseenCnt?.find((item) => item.conversation_id === id);
  const unseenCount = cnt ? cnt.unseen_message_count : 0;
  const { data: lastMessage } = useQuery({
    queryKey: ["lastMessage", message_id],
    queryFn: async () => {
      if (message_id) {
        const token = await getToken();
        if (!token) {
          throw new Error("User token not found");
        }
        return getConversationLastMessage({ message_id, token });
      }
      return null;
    },
  });
  return (
    <Link href={`/conversations/${id}`} className='w-full'>
      <Card className='p-2 flex flex-row items-center justify-between'>
        <div className='flex flex-row items-center gap-4 truncate'>
          <Avatar>
            <AvatarFallback>{name.charAt(0).toLocaleUpperCase()}</AvatarFallback>
          </Avatar>
          <div className='flex flex-col truncate'>
            <h4 className='truncate'>{name}</h4>
            {lastMessage?.sender_username && lastMessage.content ? (
              <span className='text-sm text-muted-foreground flex truncate overflow-ellipsis'>
                <p className='font-semibold'>
                  {lastMessage?.sender_username}
                  {":"}&nbsp;
                </p>
                <p className='truncate overflow-ellipsis'>{lastMessage.content}</p>
              </span>
            ) : (
              <p className='text-sm text-muted-foreground truncate'>Start the conversation!</p>
            )}
          </div>
        </div>
        {unseenCount ? <Badge>{unseenCount}</Badge> : null}
      </Card>
    </Link>
  );
}

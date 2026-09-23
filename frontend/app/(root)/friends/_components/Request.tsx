import { acceptRequest, denyRequest } from "@/api/friends";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { useAuthInfo } from "@/hooks/useAuthInfo";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Check, User, X } from "lucide-react";
import { toast } from "sonner";

type Props = {
  id: number;
  imageUrl: string;
  username: string;
  email: string;
};

export default function Request({ id, imageUrl, email, username }: Props) {
  const { token } = useAuthInfo();
  const queryClient = useQueryClient();

  const { mutate: deny, isPending: denyPending } = useMutation({
    mutationFn: () => denyRequest({ request_id: id }),
    onSuccess: () => {
      toast.success("Friend request denied");
      queryClient.invalidateQueries({ queryKey: ["friend_requests"] });
    },
    onError: () => {
      toast.error("Action failed, please try again.");
    },
  });

  const { mutate: accept, isPending: acceptPending } = useMutation({
    mutationFn: () => {
      if (!token) {
        throw new Error("Authentication is not ready");
      }
      return acceptRequest({ request_id: id, token });
    },
    onSuccess: () => {
      toast.success("Friend request accept");
      queryClient.invalidateQueries({ queryKey: ["friend_requests"] });
    },
    onError: () => {
      toast.error("Action failed, please try again.");
    },
  });

  return (
    <Card className='w-full p-2 flex flex-row items-center justify-between gap-2'>
      <div className='flex items-center gap-4 truncate'>
        <Avatar>
          <AvatarImage src={imageUrl} />
          <AvatarFallback>
            <User />
          </AvatarFallback>
        </Avatar>
        <div className='flex flex-col truncate'>
          <h4 className='truncate'>{username}</h4>
          <p className='text-xs text-muted-foreground truncate'>{email}</p>
        </div>
      </div>
      <div className='flex items-center gap-2'>
        <Button size='icon' disabled={denyPending || acceptPending} onClick={() => accept()}>
          <Check />
        </Button>
        <Button size='icon' disabled={denyPending} variant='destructive' onClick={() => deny()}>
          <X />
        </Button>
      </div>
    </Card>
  );
}

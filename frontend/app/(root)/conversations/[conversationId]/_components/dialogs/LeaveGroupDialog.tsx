import { leaveGroup } from "@/api/conversations";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useAuth } from "@clerk/nextjs";
import { useMutation } from "@tanstack/react-query";
import { Dispatch, SetStateAction } from "react";
import { toast } from "sonner";

type Props = {
  conversationId: number;
  open: boolean;
  setOpen: Dispatch<SetStateAction<boolean>>;
};

export default function LeaveGroupDialog({ conversationId, open, setOpen }: Props) {
  const { getToken } = useAuth();
  const { mutate: leave, isPending } = useMutation({
    mutationFn: async () => {
      const token = await getToken();
      if (!token) {
        throw new Error("User token not found");
      }
      return leaveGroup({ conversation_id: conversationId, token });
    },
    onSuccess: () => {
      toast.success("Group left");
    },
    onError: () => {
      toast.error("Failed to leave group");
    },
  });

  const handleLeaveGroup = async () => {
    leave();
  };

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. You will not be able to see any previous messages or send new messages to this
            group.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isPending}>Cancel</AlertDialogCancel>
          <AlertDialogAction disabled={isPending} onClick={handleLeaveGroup}>
            Leave
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

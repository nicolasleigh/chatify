import { deleteFriend } from "@/api/friends";
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

export default function RemoveFriendDialog({ conversationId, open, setOpen }: Props) {
  const { getToken } = useAuth();
  const { mutate: removeFriend, isPending } = useMutation({
    mutationFn: async () => {
      const token = await getToken();
      if (!token) {
        throw new Error("User token not found");
      }
      return deleteFriend({ conversation_id: conversationId, token });
    },
    onSuccess: () => {
      toast.success("Friend delete successfully");
    },
    onError: () => {
      toast.error("Failed to delete friend");
    },
  });

  const handleRemoveFriend = async () => {
    removeFriend();
  };

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. All messages will be deleted and you will not be able to message this user.
            All group chats will still work as normal.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isPending}>Cancel</AlertDialogCancel>
          <AlertDialogAction disabled={isPending} onClick={handleRemoveFriend}>
            Delete
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

"use client";

import { createRequest } from "@/api/friends";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useAuthInfo } from "@/hooks/useAuthInfo";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { UserPlus } from "lucide-react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

const addFriendFormSchema = z.object({
  email: z.string().min(1, { message: "This field can't be empty" }).email("Please enter a valid email"),
});

export default function AddFriendDialog() {
  const { token } = useAuthInfo();
  const { mutate: createReq, isPending } = useMutation({
    mutationFn: (email: string) => {
      if (!token) {
        throw new Error("Authentication is not ready");
      }
      return createRequest({ email, token });
    },
    onSuccess: () => {
      toast.success("Friend request sent!");
    },
    onError: (err) => {
      toast.error(err.message);
    },
  });
  const form = useForm<z.infer<typeof addFriendFormSchema>>({
    resolver: zodResolver(addFriendFormSchema),
    defaultValues: {
      email: "",
    },
  });

  const handleSubmit = async (values: z.infer<typeof addFriendFormSchema>) => {
    createReq(values.email);
  };
  return (
    <Dialog>
      <Tooltip>
        <TooltipTrigger asChild>
          <DialogTrigger asChild>
            <Button size='icon' variant='outline'>
              <UserPlus />
            </Button>
          </DialogTrigger>
        </TooltipTrigger>
        <TooltipContent>
          <p>Add Friend</p>
        </TooltipContent>
      </Tooltip>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add friend</DialogTitle>
          <DialogDescription>Send a request to connect with your friends</DialogDescription>
        </DialogHeader>

        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className='space-y-8'>
            <FormField
              control={form.control}
              name='email'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Email</FormLabel>
                  <FormControl>
                    <Input placeholder='Email...' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <Button disabled={isPending} type='submit'>
                Send
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

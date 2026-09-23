"use client";

import { createGroup } from "@/api/conversations";
import { getFriends } from "@/api/friends";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useAuthInfo } from "@/hooks/useAuthInfo";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery } from "@tanstack/react-query";
import { CirclePlus, X } from "lucide-react";
import { useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

const createGroupFormSchema = z.object({
  name: z.string().min(1, { message: "This field can't be empty" }),
  members: z.number().array().min(1, { message: "You must select at least 1 friend" }),
});

export default function CreateGroupDialog() {
  const [open, setOpen] = useState(false);
  const { userId: clerk_id, token } = useAuthInfo();
  const { data: friends } = useQuery({
    queryKey: ["friends"],
    queryFn: () => {
      if (!clerk_id || !token) {
        throw new Error("Authentication is not ready");
      }
      return getFriends({ clerk_id, token });
    },
    enabled: Boolean(clerk_id && token),
  });

  const { mutate: create, isPending } = useMutation({
    mutationFn: ({ name, member_id_arr }: { name: string; member_id_arr: number[] }) => {
      if (!clerk_id) {
        throw new Error("User ID not valid");
      }
      return createGroup({ name, member_id_arr, clerk_id });
    },
    onSuccess: () => {
      form.reset();
      toast.success("Group created");
    },
    onError: () => {
      toast.error("Failed to create group");
    },
  });

  const form = useForm<z.infer<typeof createGroupFormSchema>>({
    resolver: zodResolver(createGroupFormSchema),
    defaultValues: {
      name: "",
      members: [],
    },
  });

  const members = form.watch("members", []);

  const unselectedFriends = useMemo(() => {
    return friends ? friends.filter((friend) => !members.includes(friend.id)) : [];
  }, [members.length, friends?.length]);

  const handleSubmit = async (values: z.infer<typeof createGroupFormSchema>) => {
    await create({ name: values.name, member_id_arr: values.members });
    setOpen(false);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <Tooltip>
        <TooltipTrigger asChild>
          <DialogTrigger asChild>
            <Button size='icon' variant='outline'>
              <CirclePlus />
            </Button>
          </DialogTrigger>
        </TooltipTrigger>
        <TooltipContent>
          <p>Create Group</p>
        </TooltipContent>
      </Tooltip>

      <DialogContent className='block'>
        <DialogHeader>
          <DialogTitle>Create Group</DialogTitle>
          <DialogDescription>Add your friends to get started!</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className='space-y-8'>
            <FormField
              control={form.control}
              name='name'
              render={({ field }) => {
                return (
                  <FormItem>
                    <FormLabel>Name</FormLabel>
                    <FormControl>
                      <Input placeholder='Group name...' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                );
              }}
            />
            <FormField
              control={form.control}
              name='members'
              render={() => {
                return (
                  <FormItem>
                    <FormLabel>Name</FormLabel>
                    <FormControl>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild disabled={unselectedFriends.length === 0}>
                          <Button className='w-full' variant='outline'>
                            Select
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent className='w-full'>
                          {unselectedFriends.map((friend) => {
                            return (
                              <DropdownMenuCheckboxItem
                                key={friend.id}
                                className='flex items-center gap-2 w-full p-2'
                                onCheckedChange={(checked) => {
                                  if (checked) {
                                    form.setValue("members", [...members, friend.id]);
                                  }
                                }}
                              >
                                <Avatar className='w-8 h-8'>
                                  <AvatarImage src={friend.image_url} />
                                  <AvatarFallback>{friend.username.substring(0, 1)}</AvatarFallback>
                                </Avatar>
                                <h4 className='truncate'>{friend.username}</h4>
                              </DropdownMenuCheckboxItem>
                            );
                          })}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                );
              }}
            />
            {members && members.length ? (
              <Card className='flex items-center gap-3 overflow-x-auto w-full h-24 p-2 no-scrollbar'>
                {friends
                  ?.filter((friend) => members.includes(friend.id))
                  .map((friend) => {
                    return (
                      <div key={friend.id} className='flex flex-col items-center gap-1'>
                        <div className='relative'>
                          <Avatar>
                            <AvatarImage src={friend.image_url} />
                            <AvatarFallback>{friend.username.substring(0, 1)}</AvatarFallback>
                          </Avatar>
                          <X
                            className='text-muted-foreground w-4 h-4 absolute bottom-8 left-7 bg-muted rounded-full cursor-pointer'
                            onClick={() =>
                              form.setValue(
                                "members",
                                members.filter((id) => id !== friend.id)
                              )
                            }
                          />
                        </div>
                        <p className='truncate text-sm'>{friend.username.split(" ")[0]}</p>
                      </div>
                    );
                  })}
              </Card>
            ) : null}
            <DialogFooter>
              <Button disabled={isPending} type='submit' className='w-full'>
                Create
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

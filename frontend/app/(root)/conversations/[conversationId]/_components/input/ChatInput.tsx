"use client";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormMessage } from "@/components/ui/form";
import useConversation from "@/hooks/useConversation";
import { Message } from "@/hooks/useMessagesQuery";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import EmojiPicker, { Theme } from "emoji-picker-react";
import { SendHorizonal } from "lucide-react";
import { useTheme } from "next-themes";
import { ChangeEvent, useEffect, useRef, useState } from "react";
import { useForm } from "react-hook-form";
import TextareaAutosize from "react-textarea-autosize";
import { toast } from "sonner";
import { z } from "zod";
import MessageActionsPopover from "./MessageActionsProvider";

const chatMessageSchema = z.object({
  content: z.string().min(1, {
    message: "This field can't be empty",
  }),
});
type ChatInputParams = {
  sender_id: number;
  websocket: WebSocket | null;
};
export default function ChatInput({ sender_id, websocket }: ChatInputParams) {
  const textareaRef = useRef<HTMLTextAreaElement | null>(null);
  const emojiPickerRef = useRef<HTMLDivElement>(null);
  const [emojiPickerOpen, setEmojiPickerOpen] = useState(false);
  const [cursorPosition, setCursorPosition] = useState(0);
  const { conversationId } = useConversation();
  const { theme } = useTheme();
  const queryClient = useQueryClient();

  const { mutate: sendMsg, isPending } = useMutation({
    mutationFn: ({ type, content }: { type: string; content: string }) =>
      // createMessage({ conversation_id: parseInt(conversationId), type, content, sender_id }),
      new Promise<Message>((resolve, reject) => {
        if (!websocket || websocket.readyState !== WebSocket.OPEN) {
          reject(new Error("WebSocket is not connected"));
          return;
        }

        websocket.send(JSON.stringify({ conversation_id: parseInt(conversationId), type, content, sender_id }));
        let result: Message | undefined;
        websocket.onmessage = (event: MessageEvent<string>) => {
          result = JSON.parse(event.data) as Message;
          resolve(result);
          queryClient.setQueryData(["messages", conversationId], (old: Message[] | undefined) => [
            result,
            ...(old || []),
          ]);
        };
        setTimeout(() => {
          if (!result) reject(new Error("Message send timed out"));
        }, 2000);
      }),
    onSuccess: () => {
      form.reset();
      textareaRef.current?.focus();
    },
    onError: () => {
      toast.error("Failed to send message");
    },
  });
  const form = useForm<z.infer<typeof chatMessageSchema>>({
    resolver: zodResolver(chatMessageSchema),
    defaultValues: {
      content: "",
    },
  });

  const content = form.watch("content", "");

  const handleInputChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    const { value, selectionStart } = event.target;

    if (selectionStart !== null) {
      form.setValue("content", value);
      setCursorPosition(selectionStart);
    }
  };

  const insertEmoji = (emoji: string) => {
    const newText = [content.substring(0, cursorPosition), emoji, content.substring(cursorPosition)].join("");

    form.setValue("content", newText);

    setCursorPosition(cursorPosition + emoji.length);
  };

  const handleSubmit = async (values: z.infer<typeof chatMessageSchema>) => {
    sendMsg({ type: "text", content: values.content });
  };

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (emojiPickerRef.current && !emojiPickerRef.current.contains(event.target as Node)) {
        setEmojiPickerOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  return (
    <Card className='w-full p-2 rounded-lg relative'>
      <div className='absolute bottom-16' ref={emojiPickerRef}>
        <EmojiPicker
          open={emojiPickerOpen}
          theme={theme as Theme}
          onEmojiClick={(emojiDetails) => {
            insertEmoji(emojiDetails.emoji);
            setEmojiPickerOpen(false);
          }}
          lazyLoadEmojis
        />
      </div>
      <div className='flex gap-2 items-end w-full'>
        <MessageActionsPopover setEmojiPickerOpen={setEmojiPickerOpen} />
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className='flex gap-2 items-end w-full'>
            <FormField
              control={form.control}
              name='content'
              render={({ field }) => {
                return (
                  <FormItem className='h-full w-full'>
                    <FormControl>
                      <TextareaAutosize
                        onKeyDown={async (e) => {
                          if (e.key === "Enter" && !e.shiftKey) {
                            e.preventDefault();
                            await form.handleSubmit(handleSubmit)();
                          }
                        }}
                        rows={1}
                        maxRows={3}
                        {...field}
                        onChange={handleInputChange}
                        // onClick={handleInputChange}
                        placeholder='Type a message...'
                        className='min-h-full w-full resize-none border-0 outline-0 bg-card text-card-foreground placeholder:text-muted-foreground'
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                );
              }}
            />
            <Button disabled={isPending} size='icon' type='submit'>
              <SendHorizonal />
            </Button>
          </form>
        </Form>
      </div>
    </Card>
  );
}

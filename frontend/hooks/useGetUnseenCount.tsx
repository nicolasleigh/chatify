import { getUnseenMessageCount } from "@/api/messages";
import { useAuth } from "@clerk/nextjs";
import { useQuery } from "@tanstack/react-query";

export default function useGetUnseenCount() {
  const { getToken } = useAuth();
  const { data } = useQuery({
    queryKey: ["unseen_message_count"],
    queryFn: async () => {
      const token = await getToken();
      if (!token) {
        throw new Error("User token not found");
      }
      return getUnseenMessageCount({ token });
    },
    refetchInterval: 1000,
  });
  // const queryClient = useQueryClient();
  // return queryClient.getQueryData(["unseen_message_count"]);
  return data;
}

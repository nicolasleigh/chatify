import { getUnseenMessageCount } from "@/api/messages";
import { useAuthInfo } from "@/hooks/useAuthInfo";
import { useQuery } from "@tanstack/react-query";

export default function useGetUnseenCount() {
  const { userId: clerk_id, token } = useAuthInfo();
  const { data } = useQuery({
    queryKey: ["unseen_message_count"],
    queryFn: () => {
      if (!clerk_id || !token) {
        throw new Error("Authentication is not ready");
      }
      return getUnseenMessageCount({ clerk_id, token });
    },
    enabled: Boolean(clerk_id && token),
    refetchInterval: 1000,
  });
  // const queryClient = useQueryClient();
  // return queryClient.getQueryData(["unseen_message_count"]);
  return data;
}

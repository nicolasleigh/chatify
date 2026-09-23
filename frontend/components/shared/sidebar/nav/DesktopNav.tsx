"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { ThemeToggle } from "@/components/ui/theme/theme-toggle";
import PushNotificationButton from "@/components/shared/notifications/PushNotificationButton";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import useNavigation from "@/hooks/useNavigation";
import { UserButton } from "@clerk/nextjs";
import Link from "next/link";

export default function DesktopNav() {
  const paths = useNavigation();
  return (
    <Card className='hidden lg:flex lg:flex-col lg:justify-between lg:items-center lg:h-full lg:w-16 lg:px-2 lg:py-4'>
      <nav>
        <ul className='flex flex-col items-center gap-4'>
          {paths.map((path) => {
            return (
              <li key={path.name} className='relative'>
                <Link href={path.href}>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Button size='icon' variant={path.active ? "default" : "outline"}>
                        {path.icon}
                      </Button>
                    </TooltipTrigger>
                    {path.count ? (
                      <Badge className='absolute left-6 bottom-6 px-2 rounded-full'>{path.count}</Badge>
                    ) : null}
                    <TooltipContent>
                      <p>{path.name}</p>
                    </TooltipContent>
                  </Tooltip>
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>
      <div className='flex flex-col items-center gap-1'>
        <PushNotificationButton />
        <ThemeToggle />
        <UserButton />
      </div>
    </Card>
  );
}

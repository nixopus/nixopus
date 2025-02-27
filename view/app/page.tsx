"use client"
import { Button } from "@/components/ui/button";
import { useAppSelector } from "@/redux/hooks";
import { useRouter } from "next/navigation";

export default function Home() {
  const router = useRouter()
  const authenticated = useAppSelector(state => state.auth.isAuthenticated)

  if (authenticated) {
    router.push("/dashboard")
    return < ></>
  }

  return (
    <div className="flex h-screen flex-col items-center justify-center">
      <div className="text-5xl">Hello, Welcome to Nixopus</div>
      <div className="mt-10 flex justify-center gap-20 text-2xl">
        <Button onClick={() => router.push("/login")}>Signin</Button>
      </div>
    </div>
  );
} 
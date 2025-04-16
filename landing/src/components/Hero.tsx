"use client";
import { motion } from "framer-motion";
import { ArrowRightIcon, ChevronRight } from "lucide-react";
import Link from "next/link";
import AnimatedShinyText from "./shinyText";
import { cn } from "./lib/utils";
import { BackgroundLines } from "./ui/background-lines";
import { useEffect, useState } from "react";

function HeroMobile() {
  return (
    <div className="flex flex-col items-center space-y-6 text-center px-4">
      <div
        className={cn(
          "group rounded-full border border-white/10 bg-gradient-to-r from-indigo-500/10 to-purple-500/10 text-xs sm:text-sm text-white transition-all ease-in",
        )}
      >
        <AnimatedShinyText className="inline-flex items-center justify-center px-3 py-1 transition ease-out hover:text-indigo-300 hover:duration-300">
          <span>🚀 Built for devs. Fully open-source.</span>
        </AnimatedShinyText>
      </div>

      <div className="mb-6">
        <h1 className="text-balance text-4xl font-extrabold tracking-tight text-white">
          <span className="bg-gradient bg-clip-text text-transparent bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400">Nixopus</span>
          <br className="mb-2" />
          <span className="text-white/90">Server Management</span>
        </h1>
      </div>

      <p className="max-w-[64rem] text-balance text-sm tracking-tight text-gray-300">
        <strong className="text-indigo-300">Self-hosting</strong>, a built-in <strong className="text-purple-300">terminal emulator</strong>,{" "}
        an intuitive <strong className="text-pink-300">file manager</strong>, and{" "}
        native integrations with <strong className="text-indigo-300">Discord, Slack, Email, </strong>and more.
      </p>

      <div className="pt-4">
        <Link
          href="#waitlist"
          className={cn(
            "group relative inline-flex items-center justify-center rounded-full border bg-gradient-to-r from-indigo-500 to-purple-500 border-transparent px-5 py-2 text-sm font-semibold tracking-tighter text-white transition-all ease-in hover:from-indigo-600 hover:to-purple-600 hover:text-white focus-visible:outline-none focus-visible:ring focus-visible:ring-indigo-500 focus-visible:ring-offset-2 focus-visible:ring-offset-indigo-100",
          )}
        >
          Join Waitlist
          <ChevronRight className="ml-2 size-3 flex-shrink-0 transition-all duration-300 ease-out group-hover:translate-x-1" />
        </Link>
      </div>
    </div>
  );
}

export default function Hero() {
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768); // md breakpoint
    };
    
    checkMobile();
    window.addEventListener('resize', checkMobile);
    
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  if (isMobile) {
    return (
      <section>
        <div className="relative h-full py-6 bg-gradient-to-b from-gray-900 via-gray-950 to-black">
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_120%,rgba(120,119,198,0.1),rgba(255,255,255,0))]"></div>
          <div className="absolute inset-0 bg-[linear-gradient(to_right,#4f4f4f2e_1px,transparent_1px),linear-gradient(to_bottom,#4f4f4f2e_1px,transparent_1px)] bg-[size:14px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)]"></div>
          <div className="container z-10 flex flex-col">
            <HeroMobile />
          </div>
        </div>
      </section>
    );
  }

  return (
    <section>
      <div className="relative h-full py-10 bg-gradient-to-b from-gray-900 via-gray-950 to-black">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_120%,rgba(120,119,198,0.1),rgba(255,255,255,0))]"></div>
        <div className="absolute inset-0 bg-[linear-gradient(to_right,#4f4f4f2e_1px,transparent_1px),linear-gradient(to_bottom,#4f4f4f2e_1px,transparent_1px)] bg-[size:14px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)]"></div>
        <div className="container z-10 flex flex-col">
          <div className="grid grid-cols-1">
            <div className="flex flex-col items-center gap-6 pb-8 text-center">
              <div
                className={cn(
                  "group rounded-full border border-white/10 bg-gradient-to-r from-indigo-500/10 to-purple-500/10 text-base text-white transition-all ease-in mb-2 md:mb-0",
                )}
              >
                <AnimatedShinyText className="inline-flex items-center justify-center px-4 py-2 transition ease-out hover:text-indigo-300 hover:duration-300">
                  <span>🚀 Built for devs. Fully open-source. </span>
                </AnimatedShinyText>
              </div>
              <BackgroundLines className="relative flex flex-col gap-4 md:items-center lg:flex-row">
                <h1 className="relative mx-0 max-w-[43.5rem] text-balance pt-5 text-left text-5xl font-extrabold tracking-tight text-white dark:text-white sm:text-7xl md:mx-auto md:px-4 md:py-2 md:text-center md:text-7xl lg:text-7xl">
                  <span className="bg-gradient bg-clip-text text-transparent bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 mr-2">Nixopus</span>
                  <br />
                  All in one tool for managing your server
                </h1>
                <span className="text-neutral-90 absolute -top-3.5 left-0 z-10 rotate-3 whitespace-nowrap rounded-full bg-gradient-to-r from-indigo-500 to-purple-500 px-2.5 py-1 text-[11px] font-semibold uppercase leading-5 tracking-wide text-white md:top-12 md:-rotate-12">
                  Early Access
                </span>
              </BackgroundLines>
              <p className="max-w-[64rem] text-balance text-lg tracking-tight text-gray-300 dark:text-gray-300 md:text-xl">
                <strong className="text-indigo-300">Self-hosting</strong>, a built-in <strong className="text-purple-300">terminal emulator</strong>,{" "}
                an intuitive <strong className="text-pink-300">file manager</strong>, and{" "}
                native integrations with  <strong className="text-indigo-300">Discord, Slack, Email, </strong>and more.
                <br />
              </p>
              <div className="flex flex-col gap-4 lg:flex-row">
                <div className="flex flex-col gap-4 md:flex-row">
                  <Link
                    href="#waitlist"
                    className={cn(
                      "group relative inline-flex items-center rounded-full border bg-gradient-to-r from-indigo-500 to-purple-500 border-transparent px-6 py-3 text-base font-semibold tracking-tighter text-white transition-all ease-in hover:from-indigo-600 hover:to-purple-600 hover:text-white focus-visible:outline-none focus-visible:ring focus-visible:ring-indigo-500 focus-visible:ring-offset-2 focus-visible:ring-offset-indigo-100",
                    )}
                  >
                    Join Waitlist
                    <ChevronRight className="ml-1 size-4 flex-shrink-0 transition-all duration-300 ease-out group-hover:translate-x-1" />
                  </Link>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
